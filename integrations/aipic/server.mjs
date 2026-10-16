import { createServer } from 'node:http'
import { Readable } from 'node:stream'
import { pipeline } from 'node:stream/promises'
import { readFile, stat } from 'node:fs/promises'
import { resolve, extname, sep } from 'node:path'
import { pathToFileURL } from 'node:url'
import { handlePromptOptimizer } from './prompt-optimizer.mjs'

const mime = { '.html': 'text/html; charset=utf-8', '.js': 'text/javascript', '.css': 'text/css', '.svg': 'image/svg+xml', '.png': 'image/png', '.jpg': 'image/jpeg', '.json': 'application/json', '.webmanifest': 'application/manifest+json' }

export function createStudioServer({ origin, secret, dist }) {
  const upstream = new URL(origin)
  if (!['http:', 'https:'].includes(upstream.protocol) || upstream.username || upstream.password || upstream.pathname !== '/' || upstream.search || upstream.hash || secret.length < 32) throw new Error('Invalid MidCode gateway configuration')
  const root = resolve(dist)
  return createServer(async (req, res) => {
    res.setHeader('Referrer-Policy', 'no-referrer')
    res.setHeader('X-Content-Type-Options', 'nosniff')
    const abort = new AbortController()
    const timeout = setTimeout(() => abort.abort(), 30 * 60 * 1000)
    res.on('close', () => { clearTimeout(timeout); if (!res.writableFinished) abort.abort() })
    try {
      const pathname = new URL(req.url, 'http://localhost').pathname
      if (pathname === '/health') { res.end('ok'); return }
      if (await handlePromptOptimizer(req, res, upstream, abort.signal)) return
      if (pathname === '/midcode/exchange') {
        res.setHeader('Cache-Control', 'no-store')
        if (req.method !== 'POST') { res.writeHead(405).end(); return }
        let body = ''
        for await (const chunk of req) { body += chunk; if (body.length > 1024) { res.writeHead(413).end(); return } }
        let code
        try { code = JSON.parse(body).code } catch { res.writeHead(400).end(); return }
        if (typeof code !== 'string' || !/^[a-f0-9]{64}$/.test(code)) { res.writeHead(400).end(); return }
        const result = await fetch(new URL('/api/public/image-studio/redeem', upstream), {
          method: 'POST', headers: { 'Content-Type': 'application/json', 'X-AIPic-Handoff-Secret': secret },
          body: JSON.stringify({ code }), redirect: 'error', signal: AbortSignal.any([abort.signal, AbortSignal.timeout(10000)]),
        })
        res.writeHead(result.status, { 'Content-Type': 'application/json' })
        if (result.body) await pipeline(Readable.fromWeb(result.body), res)
        else res.end()
        return
      }
      if (pathname.startsWith('/midcode-api/')) {
        res.setHeader('Cache-Control', 'no-store')
        const route = pathname.slice('/midcode-api'.length)
        // Fixed upstream and explicit routes: this is not a caller-controlled proxy.
        if (!((req.method === 'POST' && ['/v1/images/generations', '/v1/images/edits'].includes(route)) || (req.method === 'GET' && route === '/v1/models'))) { res.writeHead(404).end(); return }
        if (!req.headers.authorization?.startsWith('Bearer ')) { res.writeHead(401).end(); return }
        const headers = { Authorization: req.headers.authorization }
        if (req.headers['content-type']) headers['Content-Type'] = req.headers['content-type']
        let requestBody = req.method === 'POST' ? req : undefined
        // Keep image results inline so the browser does not need to fetch an
        // expiring or cross-origin provider URL after the generation succeeds.
        if (req.method === 'POST' && req.headers['content-type']?.includes('application/json')) {
          let raw = ''
          for await (const chunk of req) {
            raw += chunk
            if (raw.length > 4 * 1024 * 1024) { res.writeHead(413).end(); return }
          }
          try {
            const payload = JSON.parse(raw)
            if (payload && typeof payload === 'object' && !Array.isArray(payload)) {
              payload.response_format = 'b64_json'
            }
            requestBody = JSON.stringify(payload)
            headers['Content-Type'] = 'application/json'
          } catch { res.writeHead(400).end(); return }
        }
        const result = await fetch(new URL(route, upstream), {
          method: req.method, headers, body: requestBody,
          duplex: 'half', redirect: 'error', signal: abort.signal,
        })
        res.writeHead(result.status, { 'Content-Type': result.headers.get('content-type') || 'application/json' })
        if (result.body) await pipeline(Readable.fromWeb(result.body), res)
        else res.end()
        return
      }
      if (!['GET', 'HEAD'].includes(req.method)) { res.writeHead(405).end(); return }
      let file = resolve(root, '.' + decodeURIComponent(pathname))
      if (file !== root && !file.startsWith(root + sep)) { res.writeHead(403).end(); return }
      if (!(await stat(file).catch(() => null))?.isFile()) {
        if (extname(pathname)) { res.writeHead(404).end(); return }
        file = resolve(root, 'index.html')
      }
      res.setHeader('Content-Type', mime[extname(file)] || 'application/octet-stream')
      res.setHeader('Cache-Control', extname(file) === '.html' || pathname === '/sw.js' ? 'no-cache' : 'public, max-age=3600')
      res.end(req.method === 'HEAD' ? undefined : await readFile(file))
    } catch {
      if (!res.headersSent) res.writeHead(502, { 'Content-Type': 'application/json', 'Cache-Control': 'no-store' })
      res.end('{"error":"工作台连接暂时不可用，请返回 MidCode 查看调用记录后重试"}')
    } finally { clearTimeout(timeout) }
  })
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  const server = createStudioServer({ origin: process.env.MIDCODE_ORIGIN || '', secret: process.env.AIPIC_HANDOFF_SECRET || '', dist: process.env.AIPIC_DIST || './dist' })
  server.listen(Number(process.env.PORT || 8788), process.env.HOST || '127.0.0.1')
}

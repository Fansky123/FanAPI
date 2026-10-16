import test from 'node:test'
import assert from 'node:assert/strict'
import { createServer } from 'node:http'
import { mkdtemp, writeFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { createStudioServer } from './server.mjs'

test('gateway exchanges once, forwards image uploads, and blocks other routes', async t => {
  const secret = 'test-only-shared-secret-32-characters'
  const code = 'a'.repeat(64)
  const seen = []
  let redeemed = false
  const upstream = createServer(async (req, res) => {
    let body = ''
    for await (const chunk of req) body += chunk
    seen.push({ path: req.url, body, headers: req.headers })
    res.setHeader('Content-Type', 'application/json')
    if (req.url === '/api/public/image-studio/redeem') {
      assert.equal(req.headers['x-aipic-handoff-secret'], secret)
      assert.equal(JSON.parse(body).code, code)
      if (redeemed) { res.writeHead(410).end('{}'); return }
      redeemed = true
      res.end(JSON.stringify({ apiKey: 'test-only-key', model: 'gpt-image-test' }))
    } else {
      if (req.url === '/v1/images/generations') assert.equal(JSON.parse(body).response_format, 'b64_json')
      res.end('{"data":[{"b64_json":"test"}]}')
    }
  })
  await new Promise(resolve => upstream.listen(0, '127.0.0.1', resolve))
  const dist = await mkdtemp(join(tmpdir(), 'aipic-gateway-'))
  await writeFile(join(dist, 'index.html'), '<h1>Studio</h1>')
  const gateway = createStudioServer({ origin: `http://127.0.0.1:${upstream.address().port}`, secret, dist })
  await new Promise(resolve => gateway.listen(0, '127.0.0.1', resolve))
  t.after(async () => { gateway.closeAllConnections(); upstream.closeAllConnections(); gateway.close(); upstream.close(); await rm(dist, { recursive: true }) })
  const origin = `http://127.0.0.1:${gateway.address().port}`
  const exchange = () => fetch(origin + '/midcode/exchange', { method: 'POST', body: JSON.stringify({ code }) })
  const first = await exchange()
  assert.equal(first.status, 200)
  assert.equal(first.headers.get('cache-control'), 'no-store')
  assert.equal((await first.json()).apiKey, 'test-only-key')
  assert.equal((await exchange()).status, 410)
  const multipart = '--test\r\nContent-Disposition: form-data; name="mask"; filename="mask.png"\r\nContent-Type: image/png\r\n\r\nMASK\r\n--test--\r\n'
  const image = await fetch(origin + '/midcode-api/v1/images/edits', { method: 'POST', headers: { Authorization: 'Bearer test-only-key', 'Content-Type': 'multipart/form-data; boundary=test', 'X-AIPic-Upstream': 'http://untrusted.invalid' }, body: multipart })
  assert.equal(image.status, 200)
  assert.equal(seen.at(-1).body, multipart)
  assert.equal(seen.at(-1).headers.authorization, 'Bearer test-only-key')
  assert.equal(seen.at(-1).headers['x-aipic-handoff-secret'], undefined)
  assert.equal(seen.at(-1).headers['x-aipic-upstream'], undefined)
  assert.equal((await fetch(origin + '/midcode-api/v1/images/generations', { method: 'POST' })).status, 401)
  const generation = await fetch(origin + '/midcode-api/v1/images/generations', {
    method: 'POST', headers: { Authorization: 'Bearer test-only-key', 'Content-Type': 'application/json' },
    body: JSON.stringify({ model: 'gpt-image-test', prompt: 'test' }),
  })
  assert.equal(generation.status, 200)
  assert.equal((await fetch(origin + '/midcode-api/admin/users')).status, 404)
  assert.equal((await fetch(origin + '/midcode/exchange', { method: 'POST', body: '{"code":"invalid"}' })).status, 400)
  assert.equal((await fetch(origin + '/midcode/exchange', { method: 'POST', body: 'x'.repeat(1025) })).status, 413)
  assert.equal(seen.length, 4, 'only authorized exchange and image routes reach upstream')
  assert.match(await (await fetch(origin)).text(), /Studio/)
})

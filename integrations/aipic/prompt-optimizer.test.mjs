import test from 'node:test'
import assert from 'node:assert/strict'
import { createServer } from 'node:http'
import { createStudioServer } from './server.mjs'

test('optimization scopes models to the caller key and constructs a fixed text-only request', async t => {
  const seen = []
  const upstream = createServer(async (req, res) => {
    let body = ''; for await (const chunk of req) body += chunk
    seen.push({ path: req.url, headers: req.headers, body: body ? JSON.parse(body) : null })
    res.setHeader('Content-Type', 'application/json')
    if (req.url === '/api/user/channels') res.end(JSON.stringify({ channels: [
      { type: 'llm', billing_type: 'token', routing_model: 'text-test', price_display: 'test price' },
      { type: 'llm', billing_type: 'count', routing_model: 'gpt-image-test' },
      { type: 'image', billing_type: 'image', routing_model: 'other' },
    ] }))
    else res.end(JSON.stringify({ choices: [{ finish_reason: 'stop', message: { content: '优化后的猫咪提示词' } }] }))
  })
  await new Promise(r => upstream.listen(0, '127.0.0.1', r))
  const server = createStudioServer({ origin: `http://127.0.0.1:${upstream.address().port}`, secret: 's'.repeat(32), dist: '.' })
  await new Promise(r => server.listen(0, '127.0.0.1', r))
  t.after(() => { server.closeAllConnections(); upstream.closeAllConnections(); server.close(); upstream.close() })
  const url = `http://127.0.0.1:${server.address().port}`
  const headers = { Authorization: 'Bearer caller-test-key', 'Content-Type': 'application/json' }
  assert.equal((await fetch(url + '/midcode/prompt-models')).status, 401)
  const options = await fetch(url + '/midcode/prompt-models', { headers })
  assert.deepEqual((await options.json()).models.map(x => x.id), ['text-test'])
  const result = await fetch(url + '/midcode/optimize-prompt', { method: 'POST', headers, body: JSON.stringify({ model: 'text-test', prompt: '猫咪', stream: true, messages: [], upstream: 'https://evil.invalid', max_tokens: 99999 }) })
  assert.equal(result.status, 200)
  assert.equal((await result.json()).prompt, '优化后的猫咪提示词')
  const sent = seen.at(-1)
  assert.equal(sent.path, '/v1/chat/completions')
  assert.equal(sent.headers.authorization, headers.Authorization)
  assert.equal(sent.headers['x-aipic-handoff-secret'], undefined)
  assert.equal(sent.body.stream, false)
  assert.equal(sent.body.max_tokens, 1600)
  assert.equal(sent.body.messages[0].role, 'system')
  assert.equal(sent.body.messages[1].content, '猫咪')
  assert.equal(sent.body.upstream, undefined)
  assert.equal((await fetch(url + '/midcode/optimize-prompt', { method: 'POST', headers, body: '{}' })).status, 400)
  assert.equal((await fetch(url + '/midcode/optimize-prompt', { method: 'POST', headers, body: 'x'.repeat(32769) })).status, 413)
  assert.equal(seen.length, 2)
})

import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'
import ts from 'typescript'

const source = await readFile(new URL('../../src/lib/playground-images.ts', import.meta.url), 'utf8')
const js = ts.transpileModule(source, {
  compilerOptions: { module: ts.ModuleKind.ESNext, target: ts.ScriptTarget.ES2022 },
}).outputText
const { buildImageRequest, parseImageResponse, imageProtocol, readImageFile } = await import(
  `data:text/javascript;base64,${Buffer.from(js).toString('base64')}`
)

test('all GPT image models forward 2K and 4K for generation and editing', () => {
  for (const model of ['gpt-image-2', 'gpt-image-2.5', 'gpt-image-2.5-flare', 'gpt-image-2.5-sunburst']) {
    for (const [resolution, size] of [['2K', '2560x1440'], ['4K', '3840x2160']]) {
      assert.equal(buildImageRequest(model, 'cat', resolution, '16:9').body.size, size)
      const edit = buildImageRequest(model, 'edit', resolution, '16:9', { data: 'AAA=', mimeType: 'image/png' })
      assert.equal(edit.url, '/v1/images/edits')
      assert.equal(edit.body.size, size)
    }
  }
  assert.equal(buildImageRequest('gpt-image-2.5-flare', 'cat', '4K', '16:9').body.size, '3840x2160')
  assert.equal(buildImageRequest('gpt-image-2.5-flare', 'cat', '2K', '16:9').body.size, '2560x1440')
  const request = buildImageRequest('gpt-image-2.5-sunburst', 'cat', '4K', '16:9')
  assert.equal(request.url, '/v1/images/generations')
  assert.equal(request.body.size, '3840x2160')
  assert.equal(
    buildImageRequest('gpt-image-2', 'edit', 'default', '16:9', { data: 'AAA=', mimeType: 'image/png' }).url,
    '/v1/images/edits',
  )
})

test('Gemini uses native parts, image config and inline reference image', () => {
  const request = buildImageRequest('gemini-3.1-flash-image', 'cat', '4K', '1:1', {
    data: 'AAA=',
    mimeType: 'image/png',
    src: 'data:image/png;base64,AAA=',
    name: 'reference.png',
    file: new Blob(['test']),
  })
  assert.equal(request.url, '/v1beta/models/gemini-3.1-flash-image:generateContent')
  assert.deepEqual(request.body.generationConfig.imageConfig, { aspectRatio: '1:1', imageSize: '4K' })
  assert.deepEqual(request.body.contents[0].parts[1], { inlineData: { data: 'AAA=', mimeType: 'image/png' } })
  assert.equal(imageProtocol('gemini-3.8-flash'), null)
})

test('parses both response families and rejects non-image URI schemes', () => {
  const gpt = parseImageResponse({
    data: [{ b64_json: 'AAA=' }, { url: 'https://example.test/image.png' }, { url: 'javascript:alert(1)' }],
  })
  assert.equal(gpt.images.length, 2)
  assert.equal(gpt.images[0].src, 'data:image/png;base64,AAA=')
  const gemini = parseImageResponse({
    candidates: [
      { content: { parts: [{ text: 'done' }, { inlineData: { mimeType: 'image/webp', data: 'AAA=' } }] } },
    ],
  })
  assert.equal(gemini.text, 'done')
  assert.equal(gemini.images[0].src, 'data:image/webp;base64,AAA=')
  assert.deepEqual(parseImageResponse({ error: 'no image' }).images, [])
  assert.deepEqual(parseImageResponse(null).images, [])
})

test('upload validation rejects SVG and oversized files before reading', async () => {
  await assert.rejects(readImageFile(new Blob(['<svg/>'], { type: 'image/svg+xml' })), /PNG/)
  await assert.rejects(
    readImageFile(new Blob([new Uint8Array(21 * 1024 * 1024)], { type: 'image/png' })),
    /20 MB/,
  )
})

test('Gemini Markdown images preserve signed URLs and surrounding explanation', () => {
  const url = 'https://images.example.test/photo_1.jpg?X-Amz-Credential=a%2Fauto%2Fs3&X-Amz-Signature=ABC%2B123&X-Amz-Expires=18000'
  const payload = JSON.parse(JSON.stringify({ candidates: [{ content: { parts: [
    { text: `已完成\n![原图链接5小时有效](${url})\n请及时保存。` },
    { text: `![重复](${url})` },
    { inlineData: { mimeType: 'image/png', data: 'AAA=' } },
  ] } }] }).replaceAll('&', '\\u0026'))
  const result = parseImageResponse(payload)
  assert.deepEqual(result.images, [{ src: url }, { src: 'data:image/png;base64,AAA=' }])
  assert.equal(result.text, '已完成\n\n请及时保存。')
})

test('Markdown destinations support brackets, escaped punctuation and multiple images without HTML execution', () => {
  const parse = (text) => parseImageResponse({ candidates: [{ content: { parts: [{ text }] } }] })
  assert.deepEqual(parse('![a](https://example.test/a(1).png) ![b](<https://example.test/b.png?x=1&y=2> "title")').images,
    [{ src: 'https://example.test/a(1).png' }, { src: 'https://example.test/b.png?x=1&y=2' }])
  assert.equal(parse('![a](https://example.test/a\\_1.png)').images[0].src, 'https://example.test/a_1.png')
  for (const text of ['![a](javascript:alert(1))', '![a](data:text/html;base64,AAA=)', '<img src="https://example.test/a.png">', '[普通链接](https://example.test/a.png)', '![broken](https://example.test/a.png']) {
    assert.equal(parse(text).images.length, 0)
    assert.equal(parse(text).text, text)
  }
})

import type { UserChannel } from './api/user'

export type ImageResolution = 'default' | '2K' | '4K'
export type ImageRatio = '16:9' | '1:1'
export type GeneratedImage = { src: string; width?: number; height?: number }

export function channelModel(channel?: UserChannel) {
  return channel?.routing_model || channel?.model || channel?.name || ''
}

export function imageProtocol(model: string) {
  if (/^gpt-image-/.test(model)) return 'openai'
  if (/^gemini-.*image/.test(model)) return 'gemini'
  return null
}

export function supportsImageResolution(model: string) {
  return imageProtocol(model) !== null
}

export function buildImageRequest(
  model: string,
  prompt: string,
  resolution: ImageResolution,
  ratio: ImageRatio,
  reference?: { data: string; mimeType: string },
) {
  const kind = imageProtocol(model)
  if (!kind) throw new Error('此模型尚未接入图片创作')
  const size = supportsImageResolution(model) ? resolution : 'default'
  if (kind === 'gemini') {
    return {
      url: `/v1beta/models/${encodeURIComponent(model)}:generateContent`,
      body: {
        contents: [
          {
            role: 'user',
            parts: [{ text: prompt }, ...(reference ? [{ inlineData: { data: reference.data, mimeType: reference.mimeType } }] : [])],
          },
        ],
        generationConfig: {
          responseModalities: ['TEXT', 'IMAGE'],
          imageConfig: { aspectRatio: ratio, ...(size === 'default' ? {} : { imageSize: size }) },
        },
      },
    }
  }
  return {
    url: reference ? '/v1/images/edits' : '/v1/images/generations',
    body: {
      model,
      prompt,
      n: 1,
      quality: 'high',
      output_format: 'png',
      ...(size === 'default' ? {} : { size: size === '4K' ? '3840x2160' : '2560x1440' }),
    },
  }
}

export function safeImageSource(value: unknown): value is string {
  return (
    typeof value === 'string' &&
    (/^https?:\/\//i.test(value) || /^data:image\/(png|jpeg|webp);base64,[a-z\d+/=\s]+$/i.test(value))
  )
}

// Extract Markdown image destinations without rendering upstream HTML. Keep
// signed query strings byte-for-byte (JSON decoding has already handled \u0026).
function markdownImages(text: string, add: (src: string) => void) {
  return text.replace(
    /!\[(?:\\.|[^\]\\])*\]\(\s*(<(?:https?:\/\/)[^<>\r\n]+>|https?:\/\/(?:[^\s()\\]|\\.|\([^\s()]*\))*)(?:\s+["'][^\r\n]*?["'])?\s*\)/gi,
    (markup, destination: string) => {
      const src = destination.replace(/^<|>$/g, '').replace(/\\([!"#$%&'()*+,\-./:;<=>?@[\]^_`{|}~\\])/g, '$1')
      if (!safeImageSource(src)) return markup
      add(src)
      return ''
    },
  ).trim()
}

// Treat all upstream fields as untrusted: never render arbitrary URI schemes.
export function parseImageResponse(value: unknown): { images: GeneratedImage[]; text: string } {
  const payload = value as {
    data?: { b64_json?: string; url?: string }[]
    output_format?: string
    candidates?: {
      content?: {
        parts?: {
          text?: string
          inlineData?: { data?: string; mimeType?: string }
          inline_data?: { data?: string; mime_type?: string }
        }[]
      }
    }[]
  }
  const images: GeneratedImage[] = []
  const texts: string[] = []
  const seen = new Set<string>()
  const addImage = (src: string) => {
    if (safeImageSource(src) && !seen.has(src)) {
      seen.add(src)
      images.push({ src })
    }
  }
  for (const item of Array.isArray(payload?.data) ? payload.data : []) {
    const format =
      payload.output_format === 'jpeg' ? 'jpeg' : payload.output_format === 'webp' ? 'webp' : 'png'
    const src = item.b64_json ? `data:image/${format};base64,${item.b64_json}` : item.url
    if (safeImageSource(src)) addImage(src)
  }
  for (const candidate of Array.isArray(payload?.candidates) ? payload.candidates : []) {
    for (const part of candidate.content?.parts ?? []) {
      if (typeof part.text === 'string') {
        const text = markdownImages(part.text, addImage)
        if (text) texts.push(text)
      }
      const inline =
        part.inlineData ??
        (part.inline_data ? { data: part.inline_data.data, mimeType: part.inline_data.mime_type } : undefined)
      if (inline?.data) {
        const src = `data:${inline.mimeType || 'image/png'};base64,${inline.data}`
        if (safeImageSource(src)) addImage(src)
      }
    }
  }
  return { images, text: texts.join('\n') }
}

export async function readImageFile(file: Blob): Promise<{ src: string; data: string; mimeType: string }> {
  if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type))
    throw new Error('请选择 PNG、JPEG 或 WebP 图片')
  if (file.size > 20 * 1024 * 1024) throw new Error('参考图不能超过 20 MB')
  const src = await new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result))
    reader.onerror = () => reject(new Error('读取图片失败'))
    reader.readAsDataURL(file)
  })
  return { src, data: src.slice(src.indexOf(',') + 1), mimeType: file.type }
}

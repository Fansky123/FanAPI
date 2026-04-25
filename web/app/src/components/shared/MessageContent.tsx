/**
 * MessageContent — 轻量 markdown 子集渲染器，用于对话气泡。
 *
 * 支持：
 *   - 图片：![alt](url) 或 data:image/ base64 内联图片
 *   - 引用行：以 > 开头（多模态进度提示常用此格式）
 *   - 粗体：**text**
 *   - 普通文本（保留换行）
 */

type Segment =
  | { type: 'text'; value: string }
  | { type: 'image'; alt: string; url: string }

/** 将内容字符串拆分为文字片段和图片片段。 */
function parseSegments(content: string): Segment[] {
  const segments: Segment[] = []
  // 匹配 ![alt](url) — url 允许 http/https 或 data:image/
  const imgRe = /!\[([^\]]*)\]\(((?:https?:\/\/|data:image\/)[^\)]+)\)/g
  let lastIndex = 0
  let match: RegExpExecArray | null

  while ((match = imgRe.exec(content)) !== null) {
    if (match.index > lastIndex) {
      segments.push({ type: 'text', value: content.slice(lastIndex, match.index) })
    }
    segments.push({ type: 'image', alt: match[1], url: match[2] })
    lastIndex = match.index + match[0].length
  }
  if (lastIndex < content.length) {
    segments.push({ type: 'text', value: content.slice(lastIndex) })
  }
  return segments
}

/** 渲染单行文本，处理粗体。 */
function renderInline(text: string, key: number): React.ReactNode {
  const parts = text.split(/(\*\*[^*]+\*\*)/)
  if (parts.length === 1) return text
  return (
    <span key={key}>
      {parts.map((p, i) =>
        p.startsWith('**') && p.endsWith('**') ? (
          <strong key={i}>{p.slice(2, -2)}</strong>
        ) : (
          p
        ),
      )}
    </span>
  )
}

/** 渲染纯文本片段，按行解析 blockquote。 */
function renderText(value: string, keyBase: number): React.ReactNode {
  const lines = value.split('\n')
  const nodes: React.ReactNode[] = []
  let i = 0
  for (const line of lines) {
    const stripped = line.trimStart()
    if (stripped.startsWith('> ')) {
      nodes.push(
        <span
          key={`${keyBase}-${i}`}
          className="block border-l-2 border-muted-foreground/40 pl-3 text-muted-foreground"
        >
          {renderInline(stripped.slice(2), i)}
        </span>,
      )
    } else if (stripped === '>') {
      nodes.push(<span key={`${keyBase}-${i}`} className="block h-2" />)
    } else {
      nodes.push(
        <span key={`${keyBase}-${i}`} className="block">
          {renderInline(line, i)}
        </span>,
      )
    }
    i++
  }
  return nodes
}

interface MessageContentProps {
  content: string
  /** 气泡角色：user 气泡通常不需要渲染图片 */
  role?: 'user' | 'assistant'
}

export function MessageContent({ content, role = 'assistant' }: MessageContentProps) {
  if (role === 'user') {
    return <span className="whitespace-pre-wrap">{content}</span>
  }

  const segments = parseSegments(content)

  return (
    <span className="block whitespace-pre-wrap break-words">
      {segments.map((seg, idx) => {
        if (seg.type === 'image') {
          return (
            <a key={idx} href={seg.url} target="_blank" rel="noopener noreferrer">
              <img
                src={seg.url}
                alt={seg.alt || '生成图片'}
                className="my-2 max-w-full rounded-lg border border-border"
                style={{ maxHeight: 480 }}
              />
            </a>
          )
        }
        return <span key={idx}>{renderText(seg.value, idx)}</span>
      })}
    </span>
  )
}

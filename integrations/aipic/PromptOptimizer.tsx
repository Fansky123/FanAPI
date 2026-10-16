import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { useStore } from './store'
import './prompt-optimizer.css'

type Choice = { id: string; name: string; model: string; key: string; price: string }
export default function PromptOptimizer({ onApply }: { onApply?: (value: string) => void }) {
  const prompt = useStore(s => s.prompt)
  const [open, setOpen] = useState(false)
  const [choices, setChoices] = useState<Choice[]>([])
  const [selected, setSelected] = useState('')
  const [loading, setLoading] = useState(false)
  const [busy, setBusy] = useState(false)
  const [original, setOriginal] = useState('')
  const [result, setResult] = useState('')
  const [error, setError] = useState('')
  const request = useRef<AbortController | null>(null)
  const dialog = useRef<HTMLDialogElement>(null)
  useEffect(() => useStore.subscribe((state, previous) => {
    if (state.prompt === previous.prompt) return
    request.current?.abort()
    setBusy(false); setResult(''); setError('')
  }), [])
  useEffect(() => {
    if (!open) return
    dialog.current?.showModal()
    const controller = new AbortController()
    const profiles = useStore.getState().settings.profiles
    setLoading(true); setChoices([]); setSelected(''); setError('')
    void (async () => {
      const found: Choice[] = []
      let failures = 0
      for (let i = 0; i < profiles.length; i += 4) {
        await Promise.all(profiles.slice(i, i + 4).map(async p => {
          if (!p.apiKey) return
          try {
            const response = await fetch('/midcode/prompt-models', { headers: { Authorization: `Bearer ${p.apiKey}` }, signal: AbortSignal.any([controller.signal, AbortSignal.timeout(15000)]), credentials: 'omit' })
            if (!response.ok) { failures++; return }
            const data = await response.json()
            for (const m of data.models) found.push({ id: `${p.id}:${m.id}`, name: p.name.replace('（未绑定图片模型）', ''), model: m.id, key: p.apiKey, price: m.price })
          } catch { failures++ }
        }))
        if (controller.signal.aborted) return
      }
      found.sort((a, b) => a.name.localeCompare(b.name) || a.model.localeCompare(b.model))
      setChoices(found); setSelected(found[0]?.id || ''); setLoading(false)
      if (!found.length) setError(failures ? '读取模型失败或没有可用文本模型，请重新进入创作中心后重试。' : '这些密钥尚未绑定文本模型。请在 MidCode 为一个密钥添加文本模型分组，再重新进入创作中心。')
    })()
    return () => { controller.abort(); request.current?.abort() }
  }, [open])
  const close = () => { request.current?.abort(); setBusy(false); setOpen(false); setChoices([]) }
  const optimize = async () => {
    const choice = choices.find(c => c.id === selected)
    if (!choice || busy || !prompt.trim() || prompt.length > 8000) return
    const source = prompt
    setOriginal(source)
    const controller = new AbortController(); request.current = controller
    setBusy(true); setError(''); setResult('')
    try {
      const response = await fetch('/midcode/optimize-prompt', { method: 'POST', credentials: 'omit', headers: { Authorization: `Bearer ${choice.key}`, 'Content-Type': 'application/json' }, body: JSON.stringify({ model: choice.model, prompt: source }), signal: AbortSignal.any([controller.signal, AbortSignal.timeout(95000)]) })
      const data = await response.json()
      if (!response.ok) throw new Error(data.error || '优化失败，请稍后重试。')
      if (!controller.signal.aborted && useStore.getState().prompt === source) setResult(data.prompt)
    } catch (e) { if (!controller.signal.aborted) setError(e instanceof Error && e.name !== 'TimeoutError' ? e.message : '优化超时，请重试或更换模型；原文已保留。') }
    finally { if (request.current === controller) setBusy(false) }
  }
  const changed = prompt !== original
  return <>
    <button type="button" className="midcode-optimize-trigger" disabled={!prompt.trim()} onClick={() => { setOriginal(prompt); setResult(''); setOpen(true) }}>✨ 优化提示词</button>
    {open && createPortal(<dialog ref={dialog} className="midcode-optimize-dialog" aria-labelledby="midcode-optimize-title" onCancel={close}>
      <header><h2 id="midcode-optimize-title">优化提示词</h2><button type="button" aria-label="关闭优化提示词" onClick={close}>✕</button></header>
      <p>保留你的创作意图，补充构图、光影与细节。采用后再生成图片。</p>
      <label>文本模型<select aria-label="优化使用的文本模型" value={selected} disabled={loading || busy} onChange={e => setSelected(e.target.value)}><option value="">{loading ? '正在读取可用模型…' : '请选择文本模型'}</option>{choices.map(c => <option value={c.id} key={c.id}>{c.name} · {c.model}</option>)}</select></label>
      <small>使用所选密钥，按文本模型正常计费。{choices.find(c => c.id === selected)?.price}</small>
      <label>原始提示词<textarea value={prompt} onChange={e => (onApply || useStore.getState().setPrompt)(e.target.value)} rows={4} /></label>
      <small>修改会同步到画面描述；修改后需重新优化。</small>
      {prompt.length > 8000 && <p role="alert" className="midcode-optimize-error">提示词最多 8000 字，请缩短后再优化。</p>}
      {error && <p role="alert" className="midcode-optimize-error">{error}</p>}
      {result && <label>优化结果（可编辑）<textarea aria-label="优化结果" value={result} onChange={e => setResult(e.target.value)} rows={7} /></label>}
      <footer><button type="button" onClick={close}>保留原文</button><button type="button" disabled={busy || loading || !selected || !prompt.trim() || prompt.length > 8000} onClick={() => void optimize()}>{busy ? '优化中…' : result ? '重新优化' : '开始优化'}</button>{result && <button type="button" className="midcode-optimize-primary" disabled={changed || !result.trim()} onClick={() => { if (useStore.getState().prompt !== original) return; (onApply || useStore.getState().setPrompt)(result.trim()); close() }}>采用提示词</button>}</footer>
    </dialog>, document.body)}
  </>
}

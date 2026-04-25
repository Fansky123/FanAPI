import { toast } from 'sonner'

type CopyOptions = {
  successMessage?: string
  emptyMessage?: string
  errorMessage?: string
  onSuccess?: () => void
}

function fallbackCopyText(text: string) {
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.setAttribute('readonly', 'true')
  textarea.style.position = 'fixed'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)
  textarea.select()
  const ok = document.execCommand('copy')
  document.body.removeChild(textarea)
  return ok
}

export async function copyToClipboard(text: string, options: CopyOptions = {}) {
  const {
    successMessage = '已复制',
    emptyMessage = '没有可复制的内容',
    errorMessage = '复制失败，请手动复制',
    onSuccess,
  } = options

  if (!text) {
    toast.error(emptyMessage)
    return false
  }

  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text)
    } else if (!fallbackCopyText(text)) {
      throw new Error('copy failed')
    }
    toast.success(successMessage)
    onSuccess?.()
    return true
  } catch {
    toast.error(errorMessage)
    return false
  }
}
import { readFile, writeFile, copyFile } from 'node:fs/promises'
import { resolve, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

export async function prepareOptimizer(source) {
  const here = dirname(fileURLToPath(import.meta.url))
  await copyFile(resolve(here, 'PromptOptimizer.tsx'), resolve(source, 'src/PromptOptimizer.tsx'))
  await copyFile(resolve(here, 'prompt-optimizer.css'), resolve(source, 'src/prompt-optimizer.css'))
  const patches = [
    ['RetouchWorkspace.tsx', '<div className="retouch-submit-controls">', '<PromptOptimizer onApply={value => { promptManuallyEditedRef.current = true; setPrompt(value) }} />'],
    ['RetouchSubmitBar.tsx', '<div className="retouch-submit-controls">', '<PromptOptimizer />'],
    ['InputBar.tsx', '{/* 参数 + 按钮 */}', '<PromptOptimizer />'],
  ]
  for (const [name, marker, component] of patches) {
    const path = resolve(source, 'src/components', name)
    let text = await readFile(path, 'utf8')
    if (text.includes("import PromptOptimizer from '../PromptOptimizer'")) continue
    if (!text.includes(marker)) throw new Error(`Unsupported AIPic prompt location: ${name}`)
    text = "import PromptOptimizer from '../PromptOptimizer'\n" + text.replace(marker, component + '\n' + marker)
    await writeFile(path, text)
  }
}

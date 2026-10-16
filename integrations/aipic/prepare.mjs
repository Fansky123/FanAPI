import { readFile, writeFile, copyFile } from 'node:fs/promises'
import { resolve, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'
import { prepareOptimizer } from './prepare-optimizer.mjs'
import { prepareHistory } from './prepare-history.mjs'

const source = resolve(process.argv[2] || '.')
const here = dirname(fileURLToPath(import.meta.url))
const entry = resolve(source, 'src/main.tsx')
let text = await readFile(entry, 'utf8')
if (text.includes('bootstrapMidCode')) throw new Error('Already patched')
const marker = "createRoot(document.getElementById('root')!).render("
if (!text.includes(marker)) throw new Error('Unsupported AIPic entry point')
text = "import { bootstrapMidCode, showMidCodeError } from './midcode-bootstrap'\n" + text
text = text.replace(marker, 'void bootstrapMidCode().then(() => {\n' + marker)
text += '\n}).catch(showMidCodeError)\n'
await copyFile(resolve(here, 'bootstrap.ts'), resolve(source, 'src/midcode-bootstrap.ts'))
await writeFile(entry, text)
const workspace = resolve(source, 'src/components/RetouchWorkspace.tsx')
const workspaceText = await readFile(workspace, 'utf8')
const modeMarker = "useState<RetouchGenerationMode>('edit')"
if (!workspaceText.includes(modeMarker)) throw new Error('Unsupported AIPic workspace')
await writeFile(workspace, workspaceText.replace(modeMarker, "useState<RetouchGenerationMode>('text')"))
await prepareOptimizer(source)
await prepareHistory(source)

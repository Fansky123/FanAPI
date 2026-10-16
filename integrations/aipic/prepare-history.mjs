import { readFile, writeFile, copyFile } from 'node:fs/promises'
import { resolve, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'
export async function prepareHistory(source) {
  const here = dirname(fileURLToPath(import.meta.url))
  for (const name of ['HistoryDelete.tsx', 'history-delete.css']) await copyFile(resolve(here, name), resolve(source, 'src', name))
  const path = resolve(source, 'src/components/RetouchWorkspace.tsx')
  let text = (await readFile(path, 'utf8')).replace(/\r\n/g, '\n')
  if (text.includes("import HistoryDelete from '../HistoryDelete'")) return
  const start = text.indexOf('historyTasks.map((task) => (')
  const end = text.indexOf('                ))', start)
  if (start < 0 || end < 0) throw new Error('Unsupported AIPic history list')
  let block = text.slice(start, end)
  block = block.replace('<button\n                    key={task.id}', '<div key={task.id} className="midcode-history-row">\n                  <button')
  block = block.replace('</button>', '</button>\n                  <HistoryDelete tasks={[task]} />\n                  </div>')
  text = text.slice(0, start) + block + text.slice(end)
  text = text.replace('<div className="retouch-history-list">', '<HistoryDelete tasks={historyTasks} failedOnly />\n            <div className="retouch-history-list">')
  await writeFile(path, "import HistoryDelete from '../HistoryDelete'\n" + text)
}

import { removeMultipleTasks, useStore } from './store'
import type { TaskRecord } from './types'
import './history-delete.css'

export default function HistoryDelete({ tasks, failedOnly = false }: { tasks: TaskRecord[]; failedOnly?: boolean }) {
  const candidates = tasks.filter(t => failedOnly ? t.status === 'error' : t.status !== 'running')
  if (!candidates.length) return null
  const confirm = () => {
    const ids = new Set(candidates.map(t => t.id))
    useStore.getState().setConfirmDialog({
      title: failedOnly ? '清空失败记录' : '删除历史记录',
      message: `确定删除${failedOnly ? `这 ${ids.size} 条失败记录` : '这条记录'}吗？仅清理本浏览器历史和不再使用的图片，不会删除账户调用日志或退还费用。`,
      action: async () => {
        // Recheck after confirmation; never delete a task that has been restarted.
        const current = useStore.getState().tasks.filter(t => ids.has(t.id) && (failedOnly ? t.status === 'error' : t.status !== 'running'))
        try { await removeMultipleTasks(current.map(t => t.id)) }
        catch { useStore.getState().showToast('删除失败，请刷新后重试。', 'error') }
      },
    })
  }
  return <button type="button" className="midcode-history-delete" aria-label={failedOnly ? `清空 ${candidates.length} 条失败记录` : `删除记录：${tasks[0].prompt}`} onClick={confirm}>{failedOnly ? `清空失败记录（${candidates.length}）` : '删除'}</button>
}

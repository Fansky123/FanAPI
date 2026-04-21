import { useCallback, useEffect, useLayoutEffect, useReducer, useRef, useState } from 'react'

import { getApiErrorMessage } from '@/lib/api/http'

type AsyncState<T> = { data: T; loading: boolean; error: string }
type AsyncAction<T> =
  | { type: 'start' }
  | { type: 'success'; data: T }
  | { type: 'error'; error: string }

function asyncReduce<T>(state: AsyncState<T>, action: AsyncAction<T>): AsyncState<T> {
  switch (action.type) {
    case 'start':
      return { ...state, loading: true, error: '' }
    case 'success':
      return { data: action.data, loading: false, error: '' }
    case 'error':
      return { ...state, loading: false, error: action.error }
  }
}

/**
 * General-purpose async data fetching hook.
 *
 * - Tracks loading / error / data state
 * - call `reload()` after mutations to trigger a fresh fetch
 * - Cleans up on unmount to prevent set-state on unmounted component
 */
export function useAsync<T>(fetcher: () => Promise<T>, initialData: T) {
  const [state, dispatch] = useReducer(
    (s: AsyncState<T>, a: AsyncAction<T>) => asyncReduce(s, a),
    { data: initialData, loading: true, error: '' },
  )
  const [revision, setRevision] = useState(0)

  // Sync fetcher ref after each render so async callbacks always use the latest version
  const fetcherRef = useRef(fetcher)
  useLayoutEffect(() => {
    fetcherRef.current = fetcher
  })

  useEffect(() => {
    let cancelled = false
    dispatch({ type: 'start' })

    fetcherRef.current().then(
      (result) => {
        if (!cancelled) dispatch({ type: 'success', data: result })
      },
      (err: unknown) => {
        if (!cancelled) dispatch({ type: 'error', error: getApiErrorMessage(err) })
      },
    )

    return () => {
      cancelled = true
    }
  }, [revision])

  const reload = useCallback(() => setRevision((v) => v + 1), [])

  return { data: state.data, loading: state.loading, error: state.error, reload }
}

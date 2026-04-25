import { RouterProvider } from 'react-router-dom'
import { Toaster } from 'sonner'

import { router } from '@/app/router'
import { TooltipProvider } from '@/components/ui/tooltip'

export function App() {
  return (
    <TooltipProvider>
      <RouterProvider router={router} />
      <Toaster position="top-right" richColors closeButton />
    </TooltipProvider>
  )
}

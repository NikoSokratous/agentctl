import React from 'react'
import ReactDOM from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import App from './App'
import './index.css'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
})

// Add error boundary
const rootElement = document.getElementById('root')
if (!rootElement) {
  document.body.innerHTML = '<div style="color: white; padding: 2rem;">Error: Root element not found</div>'
} else {
  try {
    ReactDOM.createRoot(rootElement).render(
      <React.StrictMode>
        <QueryClientProvider client={queryClient}>
          <BrowserRouter>
            <App />
          </BrowserRouter>
        </QueryClientProvider>
      </React.StrictMode>,
    )
  } catch (error) {
    console.error('Failed to render app:', error)
    const errorMsg = error instanceof Error ? error.message : String(error)
    document.body.innerHTML = `<div style="color: red; padding: 2rem;">Error: ${errorMsg}</div>`
  }
}

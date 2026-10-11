import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '@/app/styles.css'
import App from '@/app/App'

import "@fontsource-variable/inter";

import "@fontsource/ibm-plex-mono/400.css";

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)

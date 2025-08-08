import React from 'react'
import { createRoot } from 'react-dom/client'
import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import './styles.css'
import { AppLayout } from './shell/AppLayout'
import { Catalogue } from './pages/Catalogue'
import { Hub } from './pages/Hub'
import { Tools } from './pages/Tools'
import { VirtualServers } from './pages/VirtualServers'

const router = createBrowserRouter([
  {
    path: '/',
    element: <AppLayout />,
    children: [
      { index: true, element: <Catalogue /> },
      { path: 'hub', element: <Hub /> },
      { path: 'tools', element: <Tools /> },
      { path: 'virtual-servers', element: <VirtualServers /> },
    ],
  },
])

createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>,
)


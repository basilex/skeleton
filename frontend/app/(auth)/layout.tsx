'use client'

import { ReactNode } from 'react'
import { AuthProvider } from '@/lib/auth'
import { Toaster } from '@/components/ui/toaster'

export default function AuthLayout({
  children,
}: {
  children: ReactNode
}) {
  return (
    <AuthProvider>
      {children}
      <Toaster />
    </AuthProvider>
  )
}
'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from './context'

interface RequireAuthProps {
  children: React.ReactNode
}

export function RequireAuth({ children }: RequireAuthProps) {
  const { isAuthenticated, isLoading } = useAuth()
  const router = useRouter()

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login')
    }
  }, [isAuthenticated, isLoading, router])

  if (isLoading) {
    return (
      <div className="flex h-screen w-screen items-center justify-center">
        <div className="text-sm text-muted-foreground">Loading...</div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return null
  }

  return <>{children}</>
}

interface RequireRoleProps {
  role: string
  children: React.ReactNode
}

export function RequireRole({ role, children }: RequireRoleProps) {
  const { hasRole } = useAuth()
  
  if (!hasRole(role)) {
    return (
      <div className="flex h-screen w-screen items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-semibold">Access Denied</h1>
          <p className="text-sm text-muted-foreground mt-2">
            You don't have permission to view this page.
          </p>
        </div>
      </div>
    )
  }

  return <>{children}</>
}

interface RequirePermissionProps {
  resource: string
  action: string
  children: React.ReactNode
}

export function RequirePermission({ resource, action, children }: RequirePermissionProps) {
  const { hasPermission } = useAuth()
  
  if (!hasPermission(resource, action)) {
    return null
  }

  return <>{children}</>
}
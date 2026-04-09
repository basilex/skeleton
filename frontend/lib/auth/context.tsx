'use client'

import { createContext, useContext, useState, useEffect, ReactNode, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { apiClient } from '@/lib/api/client'
import type { User, LoginCredentials, RegisterCredentials, AuthResponse } from './types'

interface AuthContextType {
  user: User | null
  isLoading: boolean
  isAuthenticated: boolean
  login: (credentials: LoginCredentials) => Promise<void>
  register: (credentials: RegisterCredentials) => Promise<void>
  logout: () => Promise<void>
  refreshToken: () => Promise<void>
  hasPermission: (resource: string, action: string) => boolean
  hasRole: (role: string) => boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const router = useRouter()

  useEffect(() => {
    let mounted = true
    
    async function checkAuth() {
      try {
        const token = localStorage.getItem('access_token')
        if (!token) {
          if (mounted) setIsLoading(false)
          return
        }

        apiClient.setToken(token)
        const currentUser = await apiClient.get<User>('/api/v1/auth/me')
        if (mounted) setUser(currentUser)
      } catch (error) {
        console.error('Auth check failed:', error)
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        apiClient.clearToken()
      } finally {
        if (mounted) setIsLoading(false)
      }
    }
    
    checkAuth()
    
    return () => {
      mounted = false
    }
  }, []) // Empty deps - only run once

  const login = useCallback(async (credentials: LoginCredentials) => {
    const response = await apiClient.post<AuthResponse>('/api/v1/auth/login', credentials)
    
    localStorage.setItem('access_token', response.access_token)
    localStorage.setItem('refresh_token', response.refresh_token)
    apiClient.setToken(response.access_token)
    
    setUser(response.user)
    router.push('/dashboard')
  }, [router])

  const register = useCallback(async (credentials: RegisterCredentials) => {
    const response = await apiClient.post<AuthResponse>('/api/v1/auth/register', credentials)
    
    localStorage.setItem('access_token', response.access_token)
    localStorage.setItem('refresh_token', response.refresh_token)
    apiClient.setToken(response.access_token)
    
    setUser(response.user)
    router.push('/dashboard')
  }, [router])

  const logout = useCallback(async () => {
    try {
      await apiClient.post('/api/v1/auth/logout')
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      apiClient.clearToken()
      setUser(null)
      router.push('/login')
    }
  }, [router])

  const refreshToken = useCallback(async () => {
    const refreshTokenValue = localStorage.getItem('refresh_token')
    if (!refreshTokenValue) {
      throw new Error('No refresh token')
    }

    const response = await apiClient.post<AuthResponse>('/api/v1/auth/refresh', {
      refresh_token: refreshTokenValue,
    })

    localStorage.setItem('access_token', response.access_token)
    localStorage.setItem('refresh_token', response.refresh_token)
    apiClient.setToken(response.access_token)
    
    setUser(response.user)
  }, [])

  const hasPermission = useCallback((resource: string, action: string): boolean => {
    if (!user) return false
    
    // Super admin has all permissions
    if (user.permissions.includes('*:*')) return true
    
    // Check specific permission
    return user.permissions.includes(`${resource}:${action}`)
  }, [user])

  const hasRole = useCallback((role: string): boolean => {
    if (!user) return false
    return user.roles.includes(role)
  }, [user])

  return (
    <AuthContext.Provider
      value={{
        user,
        isLoading,
        isAuthenticated: !!user,
        login,
        register,
        logout,
        refreshToken,
        hasPermission,
        hasRole,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
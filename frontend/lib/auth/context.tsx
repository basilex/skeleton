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
        
        // For mock authentication, create mock user from token
        // In production, you would validate token and fetch user from /api/v1/users/{user_id}
        const mockUser: User = {
          id: '00000000-0000-0000-0000-000000000001',
          email: 'admin@skeleton.local',
          name: 'Admin User',
          roles: ['admin'],
          permissions: ['*:*'],
          active: true,
          createdAt: new Date().toISOString(),
        }
        
        if (mounted) setUser(mockUser)
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
    setIsLoading(true)
    try {
      const response = await apiClient.post<{ access_token: string; refresh_token: string }>('/api/v1/auth/login', credentials)
      
      localStorage.setItem('access_token', response.access_token)
      localStorage.setItem('refresh_token', response.refresh_token)
      apiClient.setToken(response.access_token)
      
      // For development: Mock user based on login credentials
      // In production: Fetch user profile from /api/v1/users/me with Bearer token
      // User ID is extracted from MockTokenService token format: "access-{userID}-{timestamp}"
      // MockTokenService always returns admin user with ID: 00000000-0000-0000-0000-000000000001
      const mockUser: User = {
        id: '00000000-0000-0000-0000-000000000001',
        email: credentials.email,
        name: credentials.email.split('@')[0], // Use email username as display name
        roles: ['admin'],
        permissions: ['*:*'],
        active: true,
        createdAt: new Date().toISOString(),
      }
      
      setUser(mockUser)
      setIsLoading(false)
      router.push('/dashboard')
    } catch (error) {
      setIsLoading(false)
      throw error
    }
  }, [router])

  const register = useCallback(async (credentials: RegisterCredentials) => {
    setIsLoading(true)
    try {
      const response = await apiClient.post<{ user_id: string }>('/api/v1/auth/register', credentials)
      
      // After registration, login the user
      const loginResponse = await apiClient.post<{ access_token: string; refresh_token: string }>('/api/v1/auth/login', {
        email: credentials.email,
        password: credentials.password,
      })
      
      localStorage.setItem('access_token', loginResponse.access_token)
      localStorage.setItem('refresh_token', loginResponse.refresh_token)
      apiClient.setToken(loginResponse.access_token)
      
      // For development: Mock user based on registration
      // In production: Fetch user profile from /api/v1/users/me
      const mockUser: User = {
        id: response.user_id,
        email: credentials.email,
        name: credentials.name,
        roles: ['user'],
        permissions: ['users:read', 'users:write'],
        active: true,
        createdAt: new Date().toISOString(),
      }
      
      setUser(mockUser)
      setIsLoading(false)
      router.push('/dashboard')
    } catch (error) {
      setIsLoading(false)
      throw error
    }
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
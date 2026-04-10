'use client'

import { useEffect, useState, useContext, createContext, ReactNode, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { apiClient } from '@/lib/api/client'
import type { User, LoginCredentials, RegisterCredentials, AuthResponse, UserDTO } from './types'

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

function authResponseToUser(res: AuthResponse): User {
  return {
    id: res.user_id,
    email: res.email,
    name: res.email.split('@')[0],
    roles: res.roles || [],
    permissions: [],
    active: res.is_active,
    createdAt: new Date().toISOString(),
  }
}

function userDTOToUser(dto: UserDTO): User {
  return {
    id: dto.id,
    email: dto.email,
    name: dto.email.split('@')[0],
    roles: dto.roles || [],
    permissions: [],
    active: dto.is_active,
    createdAt: dto.created_at,
  }
}

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

        const dto = await apiClient.get<UserDTO>('/api/v1/users/me')
        if (mounted) setUser(userDTOToUser(dto))
      } catch {
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
  }, [])

  const login = useCallback(async (credentials: LoginCredentials) => {
    const response = await apiClient.post<AuthResponse>('/api/v1/auth/login', credentials)

    localStorage.setItem('access_token', response.access_token)
    localStorage.setItem('refresh_token', response.refresh_token)
    apiClient.setToken(response.access_token)

    if (response.roles && response.roles.length > 0) {
      setUser(authResponseToUser(response))
    } else {
      try {
        const dto = await apiClient.get<UserDTO>('/api/v1/users/me')
        setUser(userDTOToUser(dto))
      } catch {
        setUser(authResponseToUser(response))
      }
    }

    router.push('/dashboard')
  }, [router])

  const register = useCallback(async (credentials: RegisterCredentials) => {
    await apiClient.post<{ user_id: string }>('/api/v1/auth/register', {
      email: credentials.email,
      password: credentials.password,
    })

    const loginResponse = await apiClient.post<AuthResponse>('/api/v1/auth/login', {
      email: credentials.email,
      password: credentials.password,
    })

    localStorage.setItem('access_token', loginResponse.access_token)
    localStorage.setItem('refresh_token', loginResponse.refresh_token)
    apiClient.setToken(loginResponse.access_token)

    try {
      const dto = await apiClient.get<UserDTO>('/api/v1/users/me')
      setUser(userDTOToUser(dto))
    } catch {
      setUser(authResponseToUser(loginResponse))
    }

    router.push('/dashboard')
  }, [router])

  const logout = useCallback(async () => {
    try {
      await apiClient.post('/api/v1/auth/logout')
    } catch {
      // ignore errors on logout
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

    if (response.roles && response.roles.length > 0) {
      setUser(authResponseToUser(response))
    } else {
      try {
        const dto = await apiClient.get<UserDTO>('/api/v1/users/me')
        setUser(userDTOToUser(dto))
      } catch {
        setUser(authResponseToUser(response))
      }
    }
  }, [])

  const hasPermission = useCallback((resource: string, action: string): boolean => {
    if (!user) return false
    if (user.permissions.includes('*:*')) return true
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
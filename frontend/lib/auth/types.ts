export interface User {
  id: string
  email: string
  name: string
  roles: string[]
  permissions: string[]
  active: boolean
  createdAt: string
}

export interface LoginCredentials {
  email: string
  password: string
}

export interface RegisterCredentials {
  email: string
  password: string
  name: string
}

export interface AuthResponse {
  user: User
  access_token: string
  refresh_token: string
}
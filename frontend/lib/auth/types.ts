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
  user_id: string
  email: string
  roles: string[]
  is_active: boolean
  access_token: string
  refresh_token: string
}

export interface UserDTO {
  id: string
  email: string
  roles: string[]
  is_active: boolean
  created_at: string
  updated_at: string
}
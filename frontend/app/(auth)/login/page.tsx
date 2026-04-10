'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useAuth } from '@/lib/auth'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { AuthIllustration } from '@/components/layout/auth-illustration'

export default function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const { login, isAuthenticated, isLoading: authLoading } = useAuth()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setIsLoading(true)
    try {
      await login({ email, password })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
      setIsLoading(false)
    }
  }

  if (authLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-sm text-muted-foreground">Loading...</div>
      </div>
    )
  }

  if (isAuthenticated) {
    return null
  }

  return (
    <div className="flex min-h-svh items-center justify-center p-6 md:p-10">
      <div className="flex w-full max-w-4xl overflow-hidden rounded-xl border">
        <div className="hidden md:flex md:w-1/2">
          <AuthIllustration />
        </div>
        <div className="flex w-full items-center justify-center md:w-1/2">
          <div className="w-full max-w-sm px-4">
          <div className="mb-8 flex flex-col items-center gap-2">
            <div className="flex size-8 items-center justify-center rounded-md bg-foreground text-background">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="size-4">
                <path d="M12 2v20M2 12h20" />
              </svg>
            </div>
            <h1 className="text-xl font-bold">Skeleton CRM</h1>
          </div>
          <Card>
            <CardHeader className="text-center">
              <CardTitle className="text-2xl">Sign in</CardTitle>
              <CardDescription>
                Enter your email and password to access your account
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit} className="flex flex-col gap-4">
                {error && (
                  <div className="rounded-md border border-destructive/50 bg-destructive/10 p-3">
                    <p className="text-xs text-destructive">{error}</p>
                  </div>
                )}
                <div className="flex flex-col gap-2">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="you@example.com"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                    disabled={isLoading}
                  />
                </div>
                <div className="flex flex-col gap-2">
                  <Label htmlFor="password">Password</Label>
                  <Input
                    id="password"
                    type="password"
                    placeholder="••••••••"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                    disabled={isLoading}
                  />
                </div>
                <Button type="submit" className="w-full" disabled={isLoading}>
                  {isLoading ? 'Signing in...' : 'Sign in'}
                </Button>
              </form>
              <div className="mt-4 text-center text-sm text-muted-foreground">
                Don&apos;t have an account?{' '}
                <Link href="/register" className="text-foreground underline-offset-4 hover:underline">
                  Create one
                </Link>
              </div>
            </CardContent>
          </Card>
          <div className="mt-4 text-center text-xs text-muted-foreground">
            Demo: admin@skeleton.local / Admin1234!
          </div>
        </div>
        </div>
      </div>
    </div>
  )
}
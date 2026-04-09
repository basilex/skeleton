import type { Metadata } from "next"
import { Inter, Geist } from "next/font/google"
import { cn } from "@/lib/utils"
import { AuthProvider } from "@/lib/auth"
import "./globals.css"

const geist = Geist({ subsets: ['latin'], variable: '--font-sans' })
const inter = Inter({ subsets: ["latin"] })

export const metadata: Metadata = {
  title: "Skeleton CRM",
  description: "Enterprise-grade CRM with DDD architecture",
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" className={cn("font-sans", geist.variable)}>
      <body className={inter.className}>
        <AuthProvider>
          {children}
        </AuthProvider>
      </body>
    </html>
  )
}

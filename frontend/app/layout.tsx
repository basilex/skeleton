import type { Metadata } from "next"
import { Inter, Geist } from "next/font/google"
import { cn } from "@/lib/utils";
import "./globals.css"

const geist = Geist({subsets:['latin'],variable:'--font-sans'});

const inter = Inter({ subsets: ["latin"] })

export const metadata: Metadata = {
  title: "Skeleton Business Engine",
  description: "Production-ready business engine with DDD",
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" className={cn("font-sans", geist.variable)}>
      <body className={inter.className}>{children}</body>
    </html>
  )
}

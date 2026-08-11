import type { CSSProperties, ReactNode } from "react";
import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { AppProviders } from "@/components/providers/app-providers";
import "@/styles/globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "NEXORA Admin",
  description: "NEXORA operations command center",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: ReactNode;
}>) {
  const fontVars = {
    "--nx-font-body": "var(--font-geist-sans), system-ui, sans-serif",
    "--nx-font-display": "var(--font-geist-sans), system-ui, sans-serif",
    "--nx-font-mono": "var(--font-geist-mono), ui-monospace, monospace",
  } as CSSProperties;

  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased nx-root`}
        style={fontVars}
      >
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}

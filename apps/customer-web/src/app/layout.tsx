import type { Metadata } from "next";
import { Geist } from "next/font/google";
import { AppProviders } from "@/components/providers/app-providers";
import { CustomerShell } from "@/components/shell/customer-shell";
import "@/styles/globals.css";

const geist = Geist({ subsets: ["latin"], variable: "--font-geist-sans" });

export const metadata: Metadata = {
  title: "NEXORA — Quick Commerce",
  description: "Getir-style customer web experience",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className={`${geist.variable} antialiased`}>
        <AppProviders>
          <CustomerShell>{children}</CustomerShell>
        </AppProviders>
      </body>
    </html>
  );
}

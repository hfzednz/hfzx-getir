import { AppProviders } from "@/components/providers/app-providers";
import "@/styles/globals.css";
export const metadata = { title: "NEXORA Finance" };
export default function RootLayout({ children }: { children: React.ReactNode }) {
  return <html lang="en"><body><AppProviders>{children}</AppProviders></body></html>;
}

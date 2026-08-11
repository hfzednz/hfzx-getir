"use client";

import { useState, type ReactNode } from "react";
import { QueryClientProvider } from "@tanstack/react-query";
import { createQueryClient } from "@/shared/api/query-client";
import { useChromeStore } from "@/stores/chrome-store";
import { useEffect } from "react";

function ThemeSync({ children }: { children: ReactNode }) {
  const theme = useChromeStore((s) => s.theme);

  useEffect(() => {
    const root = document.documentElement;
    root.setAttribute("data-theme", theme);
    root.classList.toggle("nx-dark", theme === "dark");
  }, [theme]);

  return <>{children}</>;
}

export function AppProviders({ children }: { children: ReactNode }) {
  const [queryClient] = useState(() => createQueryClient());

  return (
    <QueryClientProvider client={queryClient}>
      <ThemeSync>{children}</ThemeSync>
    </QueryClientProvider>
  );
}

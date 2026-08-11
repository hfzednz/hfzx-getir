"use client";

import type { ButtonHTMLAttributes, ReactNode } from "react";
import { cn } from "./tokens/cn";

export type ButtonVariant = "primary" | "secondary" | "ghost" | "danger" | "accent";
export type ButtonSize = "sm" | "md";

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  loading?: boolean;
  children?: ReactNode;
}

export function Button({
  variant = "primary",
  size = "md",
  loading = false,
  disabled,
  className,
  children,
  type = "button",
  ...props
}: ButtonProps) {
  return (
    <button
      type={type}
      disabled={disabled || loading}
      className={cn("nx-btn", `nx-btn--${variant}`, `nx-btn--${size}`, className)}
      aria-busy={loading || undefined}
      {...props}
    >
      {loading ? <span className="nx-btn__spinner" aria-hidden /> : null}
      {children}
    </button>
  );
}

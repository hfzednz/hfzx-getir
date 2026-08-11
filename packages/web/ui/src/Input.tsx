"use client";

import { forwardRef, type InputHTMLAttributes } from "react";
import { cn } from "./tokens/cn";

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  error?: boolean;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  function Input({ className, error, type = "text", ...props }, ref) {
    return (
      <input
        ref={ref}
        type={type}
        className={cn("nx-input", error && "nx-input--error", className)}
        aria-invalid={error || undefined}
        {...props}
      />
    );
  },
);

"use client";

import { forwardRef, type SelectHTMLAttributes } from "react";
import { cn } from "./tokens/cn";

export interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  error?: boolean;
}

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  function Select({ className, error, children, ...props }, ref) {
    return (
      <select
        ref={ref}
        className={cn("nx-select", error && "nx-select--error", className)}
        aria-invalid={error || undefined}
        {...props}
      >
        {children}
      </select>
    );
  },
);

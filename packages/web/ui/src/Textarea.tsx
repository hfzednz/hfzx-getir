"use client";

import { forwardRef, type TextareaHTMLAttributes } from "react";
import { cn } from "./tokens/cn";

export interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  error?: boolean;
}

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  function Textarea({ className, error, rows = 3, ...props }, ref) {
    return (
      <textarea
        ref={ref}
        rows={rows}
        className={cn("nx-textarea", error && "nx-textarea--error", className)}
        aria-invalid={error || undefined}
        {...props}
      />
    );
  },
);

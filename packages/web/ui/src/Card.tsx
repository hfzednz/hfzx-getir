import type { HTMLAttributes, ReactNode } from "react";
import { cn } from "./tokens/cn";

/**
 * Subtle bordered container for interactive surfaces only
 * (forms, selectable panels, filter groups — not decorative cards).
 */
export interface CardProps extends HTMLAttributes<HTMLDivElement> {
  children?: ReactNode;
  interactive?: boolean;
}

export function Card({
  children,
  interactive = false,
  className,
  ...props
}: CardProps) {
  return (
    <div
      className={cn("nx-card", interactive && "nx-card--interactive", className)}
      {...props}
    >
      {children}
    </div>
  );
}

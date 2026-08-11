import type { HTMLAttributes, ReactNode } from "react";
import { cn } from "./tokens/cn";

export type BadgeTone =
  | "neutral"
  | "brand"
  | "accent"
  | "success"
  | "warning"
  | "danger"
  | "info";

export interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  tone?: BadgeTone;
  children?: ReactNode;
}

export function Badge({
  tone = "neutral",
  className,
  children,
  ...props
}: BadgeProps) {
  return (
    <span className={cn("nx-badge", `nx-badge--${tone}`, className)} {...props}>
      {children}
    </span>
  );
}

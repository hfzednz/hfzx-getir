import type { HTMLAttributes } from "react";
import { cn } from "./tokens/cn";

export type SpinnerSize = "sm" | "md" | "lg";

export interface SpinnerProps extends HTMLAttributes<HTMLDivElement> {
  size?: SpinnerSize;
  label?: string;
}

export function Spinner({
  size = "md",
  label = "Loading",
  className,
  ...props
}: SpinnerProps) {
  return (
    <div
      role="status"
      aria-label={label}
      className={cn("nx-spinner", className)}
      {...props}
    >
      <span className={cn("nx-spinner__ring", `nx-spinner__ring--${size}`)} aria-hidden />
      <span className="nx-sr-only">{label}</span>
    </div>
  );
}

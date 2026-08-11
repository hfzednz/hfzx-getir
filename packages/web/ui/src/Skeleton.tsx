import type { HTMLAttributes, CSSProperties } from "react";
import { cn } from "./tokens/cn";

export interface SkeletonProps extends HTMLAttributes<HTMLDivElement> {
  width?: string | number;
  height?: string | number;
  rounded?: "sm" | "md" | "full";
}

export function Skeleton({
  width,
  height,
  rounded = "sm",
  className,
  style,
  ...props
}: SkeletonProps) {
  const radius =
    rounded === "full"
      ? "var(--nx-radius-full)"
      : rounded === "md"
        ? "var(--nx-radius-md)"
        : "var(--nx-radius-sm)";

  return (
    <div
      aria-hidden
      className={cn("nx-skeleton", className)}
      style={
        {
          width: width ?? "100%",
          height: height ?? "1em",
          borderRadius: radius,
          ...style,
        } as CSSProperties
      }
      {...props}
    />
  );
}

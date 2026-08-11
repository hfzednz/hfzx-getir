"use client";

import type { HTMLAttributes, ReactNode } from "react";
import { Button } from "./Button";
import { cn } from "./tokens/cn";

export interface BulkActionBarProps extends HTMLAttributes<HTMLDivElement> {
  selectedCount: number;
  children?: ReactNode;
  onClear?: () => void;
  clearLabel?: string;
}

export function BulkActionBar({
  selectedCount,
  children,
  onClear,
  clearLabel = "Clear",
  className,
  ...props
}: BulkActionBarProps) {
  if (selectedCount <= 0) return null;

  return (
    <div
      className={cn("nx-bulk-bar", className)}
      role="toolbar"
      aria-label="Bulk actions"
      {...props}
    >
      <span className="nx-bulk-bar__count">{selectedCount} selected</span>
      <div className="nx-bulk-bar__actions">{children}</div>
      {onClear ? (
        <Button variant="ghost" size="sm" onClick={onClear}>
          {clearLabel}
        </Button>
      ) : null}
    </div>
  );
}

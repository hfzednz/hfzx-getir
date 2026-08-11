"use client";

import type { HTMLAttributes, ReactNode } from "react";
import { Button } from "./Button";
import { cn } from "./tokens/cn";

export interface ConfirmDialogProps extends HTMLAttributes<HTMLDivElement> {
  open: boolean;
  title: string;
  description?: ReactNode;
  confirmLabel?: string;
  cancelLabel?: string;
  danger?: boolean;
  loading?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

export function ConfirmDialog({
  open,
  title,
  description,
  confirmLabel = "Confirm",
  cancelLabel = "Cancel",
  danger = false,
  loading = false,
  onConfirm,
  onCancel,
  className,
  ...props
}: ConfirmDialogProps) {
  if (!open) return null;

  return (
    <div className="nx-confirm-root" role="presentation">
      <button
        type="button"
        className="nx-confirm-backdrop"
        aria-label="Close dialog"
        onClick={onCancel}
      />
      <div
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="nx-confirm-title"
        aria-describedby={description != null ? "nx-confirm-desc" : undefined}
        className={cn("nx-confirm-dialog", className)}
        {...props}
      >
        <div className="nx-confirm-dialog__copy">
          <h2 id="nx-confirm-title" className="nx-confirm-dialog__title">
            {title}
          </h2>
          {description != null ? (
            <p id="nx-confirm-desc" className="nx-confirm-dialog__desc">
              {description}
            </p>
          ) : null}
        </div>
        <div className="nx-confirm-dialog__actions">
          <Button variant="secondary" size="sm" onClick={onCancel} disabled={loading}>
            {cancelLabel}
          </Button>
          <Button
            variant={danger ? "danger" : "primary"}
            size="sm"
            loading={loading}
            onClick={onConfirm}
          >
            {confirmLabel}
          </Button>
        </div>
      </div>
    </div>
  );
}

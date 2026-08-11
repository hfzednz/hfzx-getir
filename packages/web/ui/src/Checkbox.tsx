"use client";

import { forwardRef, type InputHTMLAttributes, type ReactNode } from "react";
import { cn } from "./tokens/cn";

export interface CheckboxProps
  extends Omit<InputHTMLAttributes<HTMLInputElement>, "type"> {
  label?: ReactNode;
}

export const Checkbox = forwardRef<HTMLInputElement, CheckboxProps>(
  function Checkbox({ className, label, id, disabled, ...props }, ref) {
    const inputId = id ?? (typeof label === "string" ? `nx-cb-${label}` : undefined);

    return (
      <label
        className={cn("nx-checkbox", disabled && "nx-checkbox--disabled", className)}
        htmlFor={inputId}
      >
        <input
          ref={ref}
          id={inputId}
          type="checkbox"
          disabled={disabled}
          className="nx-checkbox__input"
          {...props}
        />
        {label != null ? <span>{label}</span> : null}
      </label>
    );
  },
);

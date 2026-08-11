"use client";

import {
  createContext,
  useContext,
  useState,
  type HTMLAttributes,
  type ReactNode,
  type ButtonHTMLAttributes,
} from "react";
import { cn } from "./tokens/cn";

interface TabsContextValue {
  value: string;
  setValue: (v: string) => void;
}

const TabsContext = createContext<TabsContextValue | null>(null);

function useTabs() {
  const ctx = useContext(TabsContext);
  if (!ctx) throw new Error("Tabs components must be used within <Tabs>");
  return ctx;
}

export interface TabsProps extends HTMLAttributes<HTMLDivElement> {
  value?: string;
  defaultValue?: string;
  onValueChange?: (value: string) => void;
  children?: ReactNode;
}

export function Tabs({
  value: controlled,
  defaultValue = "",
  onValueChange,
  children,
  className,
  ...props
}: TabsProps) {
  const [uncontrolled, setUncontrolled] = useState(defaultValue);
  const value = controlled ?? uncontrolled;

  const setValue = (v: string) => {
    if (controlled === undefined) setUncontrolled(v);
    onValueChange?.(v);
  };

  return (
    <TabsContext.Provider value={{ value, setValue }}>
      <div className={cn("nx-tabs", className)} {...props}>
        {children}
      </div>
    </TabsContext.Provider>
  );
}

export interface TabsListProps extends HTMLAttributes<HTMLDivElement> {
  children?: ReactNode;
}

export function TabsList({ className, children, ...props }: TabsListProps) {
  return (
    <div role="tablist" className={cn("nx-tabs-list", className)} {...props}>
      {children}
    </div>
  );
}

export interface TabsTriggerProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  value: string;
  children?: ReactNode;
}

export function TabsTrigger({
  value,
  className,
  children,
  ...props
}: TabsTriggerProps) {
  const { value: active, setValue } = useTabs();
  const selected = active === value;

  return (
    <button
      type="button"
      role="tab"
      aria-selected={selected}
      tabIndex={selected ? 0 : -1}
      className={cn("nx-tabs-trigger", selected && "nx-tabs-trigger--active", className)}
      onClick={() => setValue(value)}
      {...props}
    >
      {children}
      {selected ? <span className="nx-tabs-trigger__indicator" aria-hidden /> : null}
    </button>
  );
}

export interface TabsContentProps extends HTMLAttributes<HTMLDivElement> {
  value: string;
  children?: ReactNode;
}

export function TabsContent({
  value,
  className,
  children,
  ...props
}: TabsContentProps) {
  const { value: active } = useTabs();
  if (active !== value) return null;

  return (
    <div role="tabpanel" className={cn("nx-tabs-content", className)} {...props}>
      {children}
    </div>
  );
}

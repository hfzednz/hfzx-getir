import type { Id } from "@/shared/types/common";

export type AuditSeverity = "info" | "warning" | "critical";

export interface AuditEntry {
  id: Id;
  actorId: Id;
  actorEmail: string;
  action: string;
  resource: string;
  resourceId: string;
  when: string;
  where: string;
  device: string;
  ip: string;
  sessionId: Id;
  oldValue: string | null;
  newValue: string | null;
  severity: AuditSeverity;
  sealed: boolean;
}

export interface AuditSnapshot {
  generatedAt: string;
  total: number;
  immutable: true;
  items: AuditEntry[];
}

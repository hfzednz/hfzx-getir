export interface AuditEvent {
  id: string;
  who: string;
  when: string;
  where: string;
  device: string;
  action: string;
  resource: string;
  oldValue: string;
  newValue: string;
  ip: string;
  sessionId: string;
}

export interface AuditSnapshot {
  generatedAt: string;
  events: AuditEvent[];
}

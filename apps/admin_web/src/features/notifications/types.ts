export type AlertCategory =
  | "operational"
  | "stock"
  | "security"
  | "financial"
  | "system"
  | "emergency";

export interface OpsAlert {
  id: string;
  category: AlertCategory;
  title: string;
  body: string;
  severity: "info" | "warning" | "danger";
  read: boolean;
  createdAt: string;
}

export interface NotificationsSnapshot {
  generatedAt: string;
  alerts: OpsAlert[];
}

import type { SeriesPoint } from "@/shared/lib/charts";

export interface ObsLink {
  id: string;
  kind: "logs" | "metrics" | "traces" | "dashboards";
  name: string;
  provider: string;
  url: string;
  description: string;
}

export interface ObsAlert {
  id: string;
  severity: "info" | "warning" | "critical";
  title: string;
  service: string;
  firedAt: string;
  status: "firing" | "acked" | "resolved";
}

export interface ObsIncident {
  id: string;
  severity: "sev1" | "sev2" | "sev3" | "sev4";
  title: string;
  status: "investigating" | "identified" | "mitigating" | "resolved";
  region: string;
  openedAt: string;
  commander: string;
}

export interface HealthProbe {
  id: string;
  name: string;
  kind: "http" | "tcp" | "grpc" | "synthetic";
  status: "up" | "degraded" | "down";
  latencyMs: number;
  region: string;
}

export interface SloPanel {
  id: string;
  name: string;
  sli: string;
  targetPct: number;
  currentPct: number;
  errorBudgetRemainingPct: number;
  window: string;
  status: "healthy" | "burn" | "breached";
}

export interface SlaContract {
  id: string;
  name: string;
  scope: string;
  uptimeTargetPct: number;
  currentPct: number;
  creditsAtRisk: boolean;
}

export interface ObservabilitySnapshot {
  generatedAt: string;
  kpis: {
    firingAlerts: number;
    openIncidents: number;
    probesUpPct: number;
    sloBreaches: number;
    avgErrorBudgetPct: number;
    traceIngestRps: number;
  };
  errorRateSeries: SeriesPoint[];
  latencySeries: SeriesPoint[];
  links: ObsLink[];
  alerts: ObsAlert[];
  incidents: ObsIncident[];
  health: HealthProbe[];
  slos: SloPanel[];
  slas: SlaContract[];
}

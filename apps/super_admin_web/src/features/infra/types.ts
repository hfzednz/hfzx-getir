import type { SeriesPoint } from "@/shared/lib/charts";

export type HealthStatus = "healthy" | "degraded" | "down" | "unknown";

export interface InfraCluster {
  id: string;
  name: string;
  region: string;
  provider: "aws" | "gcp" | "azure" | "onprem";
  version: string;
  nodeCount: number;
  status: HealthStatus;
  cpuPct: number;
  memPct: number;
}

export interface InfraServer {
  id: string;
  hostname: string;
  clusterId: string;
  role: "control-plane" | "worker" | "bastion" | "edge";
  instanceType: string;
  status: HealthStatus;
  cpuPct: number;
  memPct: number;
  az: string;
}

export interface K8sNamespace {
  id: string;
  name: string;
  clusterId: string;
  pods: number;
  status: HealthStatus;
  quotas: string;
}

export interface K8sDeployment {
  id: string;
  name: string;
  namespace: string;
  clusterId: string;
  replicas: number;
  ready: number;
  image: string;
  status: HealthStatus;
}

export interface K8sService {
  id: string;
  name: string;
  namespace: string;
  type: "ClusterIP" | "LoadBalancer" | "NodePort" | "ExternalName";
  clusterIp: string;
  ports: string;
}

export interface K8sIngress {
  id: string;
  name: string;
  namespace: string;
  hosts: string;
  tls: boolean;
  className: string;
  status: HealthStatus;
}

export interface Certificate {
  id: string;
  domain: string;
  issuer: string;
  expiresAt: string;
  status: "valid" | "expiring" | "expired" | "pending";
  autoRenew: boolean;
}

export interface DnsRecord {
  id: string;
  zone: string;
  name: string;
  type: "A" | "AAAA" | "CNAME" | "MX" | "TXT" | "NS";
  value: string;
  ttl: number;
  proxied: boolean;
}

export interface StorageVolume {
  id: string;
  name: string;
  type: "ebs" | "pd" | "azure-disk" | "nfs" | "s3" | "gcs";
  region: string;
  sizeGb: number;
  usedPct: number;
  status: HealthStatus;
  encrypted: boolean;
}

export interface CdnDistribution {
  id: string;
  name: string;
  provider: "cloudfront" | "cloudflare" | "fastly" | "akamai";
  domain: string;
  status: HealthStatus;
  cacheHitPct: number;
  bandwidthGbps: number;
  origins: number;
}

export interface InfraSnapshot {
  generatedAt: string;
  kpis: {
    clusters: number;
    nodes: number;
    deployments: number;
    certsExpiring: number;
    cdnHitPct: number;
    storageUsedPct: number;
  };
  cpuSeries: SeriesPoint[];
  clusters: InfraCluster[];
  servers: InfraServer[];
  namespaces: K8sNamespace[];
  deployments: K8sDeployment[];
  services: K8sService[];
  ingresses: K8sIngress[];
  certificates: Certificate[];
  dnsRecords: DnsRecord[];
  storage: StorageVolume[];
  cdn: CdnDistribution[];
}

variable "environment" { type = string }
variable "retention_days" {
  type    = number
  default = 30
}

output "stack" {
  value = {
    metrics = "prometheus + grafana"
    logs    = "loki"
    traces  = "tempo"
    otel    = "otel-collector"
    alerts  = "alertmanager"
    retention_days = var.retention_days
    environment    = var.environment
  }
}

variable "environment" { type = string }
variable "enable_managed_pg" {
  type    = bool
  default = true
}
variable "enable_redis" {
  type    = bool
  default = true
}
variable "enable_kafka" {
  type    = bool
  default = true
}
variable "backup_retention_days" {
  type    = number
  default = 35
}

output "data_plane" {
  value = {
    postgres   = var.enable_managed_pg ? "managed-ha" : "self-hosted-sts"
    redis      = var.enable_redis ? "managed" : "self-hosted"
    kafka      = var.enable_kafka ? "managed-msk-or-eq" : "strimzi"
    search     = "opensearch"
    analytics  = "clickhouse"
    object     = "s3-or-gcs-or-azure-blob"
    backup_days = var.backup_retention_days
    environment = var.environment
  }
}

variable "cluster_name" { type = string }
variable "environment" { type = string }
variable "node_min" {
  type    = number
  default = 3
}
variable "node_max" {
  type    = number
  default = 50
}
variable "enable_gpu_pool" {
  type    = bool
  default = false
}

locals {
  namespaces = [
    "nexora-system",
    "nexora-data",
    "nexora-apps",
    "nexora-obs",
    "nexora-ai",
  ]
}

output "namespaces" {
  value = local.namespaces
}

output "autoscaling" {
  value = {
    min           = var.node_min
    max           = var.node_max
    gpu_pool      = var.enable_gpu_pool
    cluster_name  = "${var.cluster_name}-${var.environment}"
  }
}

output "addons" {
  value = [
    "metrics-server",
    "cluster-autoscaler",
    "aws-ebs-csi / gcp-pd / azure-disk",
    "external-dns",
    "cert-manager",
    "ingress-nginx or istio-gateway",
  ]
}

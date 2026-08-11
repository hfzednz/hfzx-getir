# NEXORA network module — multi-cloud ready interface (AWS-shaped defaults).

variable "name" {
  type = string
}

variable "cidr" {
  type    = string
  default = "10.40.0.0/16"
}

variable "az_count" {
  type    = number
  default = 3
}

locals {
  public_subnets  = [for i in range(var.az_count) : cidrsubnet(var.cidr, 4, i)]
  private_subnets = [for i in range(var.az_count) : cidrsubnet(var.cidr, 4, i + var.az_count)]
  data_subnets    = [for i in range(var.az_count) : cidrsubnet(var.cidr, 4, i + var.az_count * 2)]
}

output "vpc_cidr" {
  value = var.cidr
}

output "public_subnets" {
  value = local.public_subnets
}

output "private_subnets" {
  value = local.private_subnets
}

output "data_subnets" {
  value = local.data_subnets
}

output "tags" {
  value = {
    Project     = "nexora"
    Component   = "network"
    ManagedBy   = "terraform"
    Name        = var.name
  }
}

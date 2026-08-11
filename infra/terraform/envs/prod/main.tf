module "network" {
  source = "../../modules/network"
  name   = "nexora-prod"
  cidr   = "10.50.0.0/16"
}

module "kubernetes" {
  source         = "../../modules/kubernetes"
  cluster_name   = "nexora"
  environment    = "prod"
  node_min       = 6
  node_max       = 80
  enable_gpu_pool = true
}

module "data" {
  source                = "../../modules/data"
  environment           = "prod"
  backup_retention_days = 35
}

module "observability" {
  source         = "../../modules/observability"
  environment    = "prod"
  retention_days = 45
}

output "summary" {
  value = {
    network       = module.network.tags
    cluster       = module.kubernetes.autoscaling
    namespaces    = module.kubernetes.namespaces
    data          = module.data.data_plane
    observability = module.observability.stack
  }
}

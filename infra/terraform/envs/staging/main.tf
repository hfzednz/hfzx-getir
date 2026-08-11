module "network" {
  source = "../../modules/network"
  name   = "nexora-staging"
  cidr   = "10.40.0.0/16"
}

module "kubernetes" {
  source       = "../../modules/kubernetes"
  cluster_name = "nexora"
  environment  = "staging"
  node_min     = 3
  node_max     = 20
}

module "data" {
  source      = "../../modules/data"
  environment = "staging"
}

module "observability" {
  source      = "../../modules/observability"
  environment = "staging"
}

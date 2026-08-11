terraform {
  required_version = ">= 1.6.0"
  required_providers {
    kubernetes = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
  }
}

variable "environment" {
  type        = string
  description = "dev|staging|prod"
}

variable "region" {
  type    = string
  default = "eu-central-1"
}

variable "cloud" {
  type        = string
  description = "aws|gcp|azure|hybrid"
  default     = "aws"
}

variable "cluster_name" {
  type    = string
  default = "nexora"
}

output "cluster_name" {
  value = "${var.cluster_name}-${var.environment}"
}

output "environment" {
  value = var.environment
}

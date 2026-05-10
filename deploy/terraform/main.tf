terraform {
  required_version = ">= 1.7"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }

  backend "s3" {
    # Configure via -backend-config or env vars at init time.
    # terraform init -backend-config="bucket=<state-bucket>" \
    #               -backend-config="key=live-rack/terraform.tfstate" \
    #               -backend-config="region=us-east-1"
    encrypt = true
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "live-rack"
      Environment = var.env
      ManagedBy   = "terraform"
    }
  }
}

locals {
  name_prefix = "lr-${var.env}"
}

module "network" {
  source = "./modules/network"

  name_prefix        = local.name_prefix
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
}

module "db" {
  source = "./modules/db"

  name_prefix        = local.name_prefix
  env                = var.env
  vpc_id             = module.network.vpc_id
  private_subnet_ids = module.network.private_subnet_ids
  db_instance_class  = var.db_instance_class
  db_storage_gb      = var.db_storage_gb
}

module "cache" {
  source = "./modules/cache"

  name_prefix        = local.name_prefix
  vpc_id             = module.network.vpc_id
  private_subnet_ids = module.network.private_subnet_ids
  cache_node_type    = var.cache_node_type
}

module "bus" {
  source = "./modules/bus"

  name_prefix        = local.name_prefix
  env                = var.env
  vpc_id             = module.network.vpc_id
  private_subnet_ids = module.network.private_subnet_ids
}

module "observability" {
  source = "./modules/observability"

  name_prefix        = local.name_prefix
  env                = var.env
  vpc_id             = module.network.vpc_id
  private_subnet_ids = module.network.private_subnet_ids
  alert_email        = var.alert_email
}

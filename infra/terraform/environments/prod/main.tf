# environments/prod/main.tf - AWS Real
/*
terraform {
  required_version = ">= 1.10"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # Producción usa S3 backend con native locking
  backend "s3" {
    bucket       = "sql-to-nosql-terraform-state"
    key          = "prod/terraform.tfstate"
    region       = "us-east-1"
    encrypt      = true
    use_lockfile = true
  }
}

provider "aws" {
  region  = var.aws_region
  profile = var.aws_profile

  default_tags {
    tags = {
      Project     = var.project
      Environment = "prod"
      ManagedBy   = "Terraform"
    }
  }
}

locals {
  name_prefix = "${var.project}-${var.environment}"
}

# ... tus módulos ...
*/

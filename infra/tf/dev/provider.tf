provider "aws" {
  region = var.aws_region
}

terraform {
  backend "s3" {
    bucket         = "mail-service-terraform-state-dev"
    key            = "mail-service-dev/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "mail-service-terraform-state-lock-dev"
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

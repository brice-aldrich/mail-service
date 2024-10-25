provider "aws" {
  region = var.aws_region
}

terraform {
  backend "s3" {
    bucket         = "mail-service-terraform-state-prod"
    key            = "mail-service-prod/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "mail-service-terraform-state-lock-prod"
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

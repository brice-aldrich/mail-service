data "terraform_remote_state" "website" {
  backend = "s3"
  
  config = {
    bucket = "my-website-terraform-state-prod"
    key    = "my-website-prod/terraform.tfstate"
    region = "us-east-1"
  }
}
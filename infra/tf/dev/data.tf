data "terraform_remote_state" "website" {
  backend = "s3"
  
  config = {
    bucket = "my-website-terraform-state-dev"
    key    = "my-website-dev/terraform.tfstate"
    region = "us-east-1"
  }
}
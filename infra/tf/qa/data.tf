data "terraform_remote_state" "website" {
  backend = "s3"
  
  config = {
    bucket = "my-website-terraform-state-qa"
    key    = "my-website-qa/terraform.tfstate"
    region = "us-east-1"
  }
}
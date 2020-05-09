terraform {
  backend "s3" {
    bucket = "sams-terraform"
    key    = "samsverynice/terraform.tfstate"
    region = "us-east-1"
  }
}

provider "aws" {
  version = "~> 2.0"
  region  = "us-east-2"
}

provider "aws" {
  alias  = "us-east-1"
  region = "us-east-1"
}

provider "google" {
  project = var.gcloud_project_id
  region  = "us-central1"
}

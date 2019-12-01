provider "aws" {
  version = "~> 2.0"
  region  = "us-east-2"
  profile = "nice"
}

provider "aws" {
  alias   = "us-east-1"
  region  = "us-east-1"
  profile = "nice"
}

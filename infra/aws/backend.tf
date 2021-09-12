terraform {
  backend "s3" {
    bucket = "helloworld-tfstate"
    key    = "terraform.tfstate"
    dynamodb_table = "helloworld-tfstate-lock"
    region = "ap-southeast-1"
  }
}

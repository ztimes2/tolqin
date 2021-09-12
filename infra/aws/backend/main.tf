provider "aws" {
  region = "ap-southeast-1"
}

resource "aws_s3_bucket" "tfstate_bucket" {
 bucket = "helloworld-tfstate"
 acl    = "private"

 versioning {
   enabled = true
 }
}

resource "aws_s3_bucket_public_access_block" "tfstate_bucket_block" {
 bucket = aws_s3_bucket.tfstate_bucket.id

 block_public_acls       = true
 block_public_policy     = true
 ignore_public_acls      = true
 restrict_public_buckets = true
}

resource "aws_dynamodb_table" "tfstate_lock_table" {
 name           = "helloworld-tfstate-lock"
 read_capacity  = 1
 write_capacity = 1
 hash_key       = "LockID"

 attribute {
   name = "LockID"
   type = "S"
 }
}
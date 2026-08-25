terraform {
  required_providers {
    minio = {
      source = "aminueza/minio"
      version = "3.40.1"
    }
  }
}

variable bucket_name {
  type = string
}

resource "minio_s3_bucket" "state_terraform_s3" {
    bucket = "${var.bucket_name}"
    acl    = "public"
}

output "bucket_id" {
    value = "${minio_s3_bucket.state_terraform_s3.id}"
}

output "bucket_url" {
    value = "${minio_s3_bucket.state_terraform_s3.bucket_domain_name}"
}
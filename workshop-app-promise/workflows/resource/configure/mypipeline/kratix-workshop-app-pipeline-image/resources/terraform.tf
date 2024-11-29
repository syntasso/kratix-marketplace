terraform {
  required_providers {
    minio = {
      source = "aminueza/minio"
      version = "2.0.1"
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

output "minio_id" {
    value = "${minio_s3_bucket.state_terraform_s3.id}"
}

output "minio_url" {
    value = "${minio_s3_bucket.state_terraform_s3.bucket_domain_name}"
}
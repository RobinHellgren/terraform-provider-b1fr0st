variable "project_id" {
  type        = string
  description = "Project UUID that will own the bucket."
}

resource "b1fr0st_bucket" "assets" {
  project_id = var.project_id
  name       = "project-assets"
  quota_gib  = 100
}

output "s3_endpoint" {
  value = b1fr0st_bucket.assets.s3_endpoint
}

output "access_key_id" {
  value = b1fr0st_bucket.assets.access_key_id
}

output "secret_access_key" {
  value     = b1fr0st_bucket.assets.secret_access_key
  sensitive = true
}

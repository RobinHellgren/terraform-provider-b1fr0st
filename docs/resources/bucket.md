# b1fr0st_bucket (Resource)

Creates and manages a private Garage S3 bucket with a dedicated access key and an adjustable size quota. A custom public hostname is optional and uses the tenant's Pangolin domain and an online Newt site.

The generated secret access key is returned once by the API and stored in Terraform state. Use encrypted remote state and keep state access restricted.

## Example Usage

```terraform
resource "b1fr0st_bucket" "assets" {
  project_id = var.project_id
  name       = "project-assets"
  quota_gib  = 100
}
```

Optional public access:

```terraform
resource "b1fr0st_bucket" "site" {
  project_id = var.project_id
  name       = "project-site"
  quota_gib  = 25
  domain_id  = var.domain_id
  subdomain  = "objects"
  site_id    = var.newt_site_id
}
```

## Argument Reference

- `project_id` (Required) — UUID of the project that owns the bucket.
- `name` (Required) — Globally unique Garage bucket name.
- `quota_gib` (Required) — Maximum bucket data size. The sum of bucket quotas must fit within the project's storage quota.
- `domain_id` (Optional) — Pangolin domain assigned to the tenant.
- `subdomain` (Optional) — DNS label for the public hostname. Requires `domain_id` and `site_id`.
- `site_id` (Optional) — Online Newt site that routes traffic to Garage. Requires `domain_id` and `subdomain`.

Changing domain settings in Terraform replaces the bucket resource. In the console, domain assignment can be edited in place.

## Attribute Reference

- `id` — Platform bucket ID.
- `hostname` — Public HTTPS hostname when configured.
- `s3_endpoint` — S3-compatible endpoint.
- `access_key_id` — Dedicated Garage access key ID.
- `secret_access_key` (Sensitive) — Generated secret, available only from the create response.
- `used_bytes` — Current data usage.
- `objects` — Current object count.

## Import

Buckets can be imported by platform bucket ID:

```shell
terraform import b1fr0st_bucket.assets <bucket-id>
```

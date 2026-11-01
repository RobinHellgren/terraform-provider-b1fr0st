# b1fr0st_disk

Creates a project-owned custom disk on the shared TrueNAS-backed Incus storage pool. Disks are detached volumes; attach one to a workload through its configuration before mounting it in the guest. Terraform deletion permanently removes disk data and is rejected while the disk is attached. Increasing `size_gib` grows the disk; shrinking is rejected to protect data.

```terraform
resource "b1fr0st_disk" "data" {
  project_id = var.project_id
  name       = "app-data"
  size_gib   = 40
}
```

## Import

Import by the opaque disk ID shown by the platform API:

```shell
terraform import b1fr0st_disk.data '<disk-id>'
```

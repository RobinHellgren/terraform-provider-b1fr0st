# b1fr0st_instance

Creates an Incus container or virtual machine in a tenant project.

## Arguments

- `tenant_id` (Required) Tenant UUID.
- `name` (Required) Instance name.
- `type` (Optional, defaults to `container`) `container` or `virtual-machine`.
- `image` (Required) Supported image alias.
- `profile` (Optional, defaults to `micro`) One of `micro`, `small`, `medium`, `large`, `xlarge`.
- `cloud_init` (Optional, Sensitive) Cloud-init user data. Requires a supported `/cloud` image.

## Attributes

- `id` Platform ID (`tenant-id:name`).
- `project` Incus project assigned to the tenant.

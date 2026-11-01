# b1fr0st_instance

Creates an Incus container or virtual machine in a tenant project. Administrators can target a project directly for system services.

## Arguments

- `project_id` (Required) Project UUID. Tenant accounts own projects; administrators can also manage workloads in the system project.
- `network` (Optional) Managed Incus bridge to attach to the workload, such as `incusbr0` for an administrator-managed system service.
- `name` (Required) Instance name.
- `type` (Optional, defaults to `container`) `container` or `virtual-machine`.
- `image` (Required) Supported image alias.
- `profile` (Optional, defaults to `micro`) One of `micro`, `small`, `medium`, `large`, `xlarge`.
- `cloud_init` (Optional, Sensitive) Cloud-init user data. Requires a supported `/cloud` image.

## Attributes

- `id` Platform ID (`tenant-id:name`).
- `project` Incus project assigned to the tenant.

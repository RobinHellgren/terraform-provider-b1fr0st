# b1fr0st_instance

Creates a managed container, virtual machine, or OCI application in an Incus project. Tenant accounts own projects; administrator services can use the system project. Tenant networking uses the project's isolated OVN overlay. Instances can consume persistent NAS-backed paths and receive cloud-init user data.

## Arguments

- `project_id` (Required) Project UUID that groups the workload.
- `network` (Optional) Managed OVN network to attach to the instance. Changing it replaces the instance.
- `name` (Required) Instance name.
- `type` (Optional, defaults to `container`) One of `container`, `virtual-machine`, or `application`.
- `image` (Required) Supported image alias or GHCR image reference for an application.
- `image_secret` (Optional, Sensitive) Private GHCR credentials in `username:token` form.
- `profile` (Optional, defaults to `micro`) One of `micro`, `small`, `medium`, `large`, or `xlarge`.
- `cloud_init` (Optional, Sensitive) Cloud-init document for supported cloud images.
- `environment` (Optional, Sensitive) Environment values for OCI applications.
- `persistent_paths` (Optional) Container paths backed by persistent volumes on the shared NAS pool.

## Attributes

- `id` Platform instance identifier.
- `project` Incus project name.

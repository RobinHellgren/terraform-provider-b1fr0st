# b1fr0st_ingress

Creates a managed Pangolin HTTP resource, enables its TLS certificate, and routes it through the selected online Newt site.

Ingress owns the complete connection path. For workloads attached to OVN, the platform allocates a reserved uplink address, creates an Incus network forward to the workload, and targets Pangolin at that forward. Destroying the ingress removes the Pangolin resource and the Incus forward. Other network types continue to use the platform-managed host proxy. The OVN uplink and Newt routing must be configured for `10.0.2.0–10.0.99.255` before OVN ingress can be created.

## Arguments

- `project_id` (Required) Project UUID that owns the workload and ingress.
- `name` (Required) Pangolin resource display name.
- `domain_id` (Required) Domain assigned to the tenant.
- `subdomain` (Required) DNS label under the selected domain.
- `site_id` (Required) Online Newt site identifier.
- `instance_name` (Optional) Running workload name. Admins may omit this for system console ingress.
- `port` (Required) Service port.

## Attributes

- `id` Pangolin resource identifier.
- `hostname` Public hostname.
- `certificate` Certificate status.
- `target` Resolved upstream target.

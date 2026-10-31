# b1fr0st_ingress

Creates a managed Pangolin HTTP resource with SSL enabled and routes it through the selected online Newt site. Pangolin owns certificate issuance and renewal.

## Arguments

- `tenant_id` (Required) Tenant UUID.
- `name` (Required) Pangolin resource display name.
- `domain_id` (Required) Pangolin domain assigned to the tenant.
- `subdomain` (Required) DNS label.
- `site_id` (Required) Online Newt site ID.
- `instance_name` (Optional) Running instance name. Administrators may omit this for system console ingress.
- `port` (Required) Service port. System console ingress requires port `8080`.

## Attributes

- `id` Pangolin resource identifier.
- `hostname` Public hostname.
- `certificate` Certificate management status.
- `target` Resolved upstream target.

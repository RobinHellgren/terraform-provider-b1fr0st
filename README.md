# Terraform provider for b1fr0st.dev

Manage Incus workloads and Pangolin ingress from Terraform through the b1fr0st.dev API.

```hcl
terraform {
  required_providers {
    b1fr0st = {
      source  = "robinhellgren/b1fr0st"
      version = "~> 0.3.0"
    }
  }
}

provider "b1fr0st" {}

resource "b1fr0st_instance" "app" {
  project_id = var.project_id
  name      = "app"
  type      = "container"
  image     = "ubuntu/24.04"
  profile   = "small"
}
```

Set `B1FR0ST_API_TOKEN` in the environment. Tokens are created in the Cloud console under **Settings → API access**. You can set `api_url` for a self-hosted API; it defaults to `https://cloud.b1fr0st.dev/api/v1`.

## Resources

- `b1fr0st_instance` manages a container, VM, or GHCR OCI application in a project. Tenant accounts own projects. Administrators can target the system project for platform services. Tenant networking uses the project's OVN overlay. OCI workloads accept sensitive environment values and private registry pull tokens. `persistent_paths` provisions paths on the shared NAS pool that survive instance replacement and deletion.
- `b1fr0st_network` creates an isolated VXLAN-backed OVN network for a project. Incus chooses an unused private subnet and configures DHCP/NAT. OVN ACLs can filter traffic between workloads. The platform refuses to delete a network while workloads use it.
- `b1fr0st_network_forward` forwards one reserved uplink address and port to a workload on an OVN network. Listen addresses are routed by its uplink; the platform reserves `10.0.2.0–10.0.99.255` for Pangolin forwards.
- `b1fr0st_disk` creates a detached custom disk in the shared TrueNAS-backed Incus storage pool. Terraform can grow disks; shrinking and deleting an attached disk are rejected to protect data.
- `b1fr0st_bucket` creates a private Garage S3 bucket, a dedicated access key, and an adjustable size quota. Optionally configure a tenant-owned Pangolin domain and Newt site for HTTPS access. The secret access key is returned only at creation and is stored in Terraform state, so protect that state as a secret.
- VM bootstrapping uses the `cloud_init` field with cloud-enabled Ubuntu or Debian images. Build the document with Terraform's `yamlencode` so users and packages are ordinary HCL values rather than a template full of shell escaping.
- `b1fr0st_ingress` creates a Pangolin resource, enables its certificate, and connects it through an online Newt site. For OVN workloads, ingress automatically owns the Incus network forward and allocated uplink address; for other network types, it uses the managed host proxy. The OVN uplink and Newt route must allow `10.0.2.0–10.0.99.255`. For tenant ingress, provide `instance_name`; admins can omit it to publish the system console on port 8080. The tenant must have the domain assigned, and the selected site must be online. Changing ingress settings replaces the resource.

See [`examples`](examples) for complete configurations. Terraform state can contain cloud-init user data, environment values, and registry credentials; protect state as a secret even though the provider marks those attributes sensitive.

## Terraform Registry installation

The repository is laid out for `registry.terraform.io/robinhellgren/b1fr0st`. After installing the release signing key as a verified GPG key in the Terraform Registry account, add the GitHub repository secrets `GPG_PRIVATE_KEY`, `GPG_FINGERPRINT`, and `GPG_PASSPHRASE`. Publishing a GitHub release tagged `vX.Y.Z` then runs GoReleaser to build platform archives, checksums, and the detached checksum signature required for Registry releases. The first Registry publication must be initiated in the Registry UI by the namespace owner.

Until Registry publication is enabled, build locally with `go build -o terraform-provider-b1fr0st .`; Terraform's `dev_overrides` can point to the resulting binary. See HashiCorp's [provider development overrides](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers).

## Development

Requires Go 1.25+. Build with `go build ./...`. Release archives are built from annotated Git tags by the included GitHub Actions workflow.

# Terraform provider for b1fr0st.dev

Manage Incus workloads and Pangolin ingress from Terraform through the b1fr0st.dev API.

```hcl
terraform {
  required_providers {
    b1fr0st = {
      source  = "robinhellgren/b1fr0st"
      version = "~> 0.1"
    }
  }
}

provider "b1fr0st" {}

resource "b1fr0st_instance" "app" {
  tenant_id = var.tenant_id
  name      = "app"
  type      = "container"
  image     = "ubuntu/24.04"
  profile   = "small"
}
```

Set `B1FR0ST_API_TOKEN` in the environment. Tokens are created in the Cloud console under **Settings → API access**. You can set `api_url` for a self-hosted API; it defaults to `https://cloud.b1fr0st.dev/api/v1`.

## Resources

- `b1fr0st_instance` manages a container, VM, or GHCR OCI application. Set exactly one of `tenant_id` or `project_id`; administrators can target the system project for platform services. The optional `network` attaches a managed bridge. OCI workloads accept sensitive environment values and private registry pull tokens. `persistent_paths` provisions paths on the shared NAS pool that survive instance replacement and deletion.
- VM bootstrapping uses the `cloud_init` field with cloud-enabled Ubuntu or Debian images. Build the document with Terraform's `yamlencode` so users and packages are ordinary HCL values rather than a template full of shell escaping.
- `b1fr0st_ingress` manages tenant Pangolin routes and TLS through the platform API.
- `b1fr0st_ingress` creates a Pangolin resource, enables its certificate, and connects it through an online Newt site. For tenant ingress, provide `instance_name`; admins can omit it to publish the system console on port 8080. The tenant must have the domain assigned, and the selected site must be online. Changing ingress settings replaces the resource.

See [`examples`](examples) for complete configurations. Terraform state can contain cloud-init user data, environment values, and registry credentials; protect state as a secret even though the provider marks those attributes sensitive.

## Terraform Registry installation

The repository is laid out for `registry.terraform.io/robinhellgren/b1fr0st`. After installing the release signing key as a verified GPG key in the Terraform Registry account, add the GitHub repository secrets `GPG_PRIVATE_KEY`, `GPG_FINGERPRINT`, and `GPG_PASSPHRASE`. Publishing a GitHub release tagged `vX.Y.Z` then runs GoReleaser to build platform archives, checksums, and the detached checksum signature required for Registry releases. The first Registry publication must be initiated in the Registry UI by the namespace owner.

Until Registry publication is enabled, build locally with `go build -o terraform-provider-b1fr0st .`; Terraform's `dev_overrides` can point to the resulting binary. See HashiCorp's [provider development overrides](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers).

## Development

Requires Go 1.25+. Build with `go build ./...`. Release archives are built from annotated Git tags by the included GitHub Actions workflow.

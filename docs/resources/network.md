# b1fr0st_network

Creates an isolated VXLAN-backed OVN network and assigns it to an Incus project. The platform selects a non-overlapping private IPv4 subnet and configures NAT/DHCP. OVN ACLs can filter traffic between instances on the overlay. The network cannot be deleted while workloads are attached.

```terraform
resource "b1fr0st_network" "apps" {
  project_id = var.project_id
  name       = "apps"
}
```

## Import

Import by network name:

```shell
terraform import b1fr0st_network.apps apps
```

`network_type` is computed from the platform (new resources report `vxlan`). The `config`, `managed`, `project`, and Incus `type` attributes are also read from the platform. Use the console to edit supported network settings and ACL rules.

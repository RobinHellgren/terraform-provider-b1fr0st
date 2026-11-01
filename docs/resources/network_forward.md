# b1fr0st_network_forward (Resource)

Creates an Incus network forward on a managed OVN network. The listen address must be allowed and reachable through the configured uplink. This resource owns one listen address and one protocol/port mapping.

```terraform
resource "b1fr0st_network_forward" "web" {
  project_id     = var.project_id
  network_name   = "tenant-apps"
  listen_address = "192.0.2.20"
  listen_port    = 443
  target_address = "10.250.1.20"
  target_port    = 8443
  protocol       = "tcp"
  description    = "Tenant web ingress"
}
```

The `listen_address` must be reserved for Incus forwarding and reachable from the intended ingress path. The provider does not allocate addresses from the underlay pool.

## Required arguments

* `project_id` - Project UUID that owns the network.
* `network_name` - Managed Incus OVN network.
* `listen_address` - Uplink address Incus will accept connections on.
* `listen_port` - Exposed TCP or UDP port.
* `target_address` - IPv4 address of a workload on the network.
* `target_port` - Workload port.
* `protocol` - `tcp` or `udp`.

## Optional arguments

* `description` - Description displayed by Incus.

All configuration changes replace this resource. To expose multiple ports on one listen address, use Incus directly until the provider adds a grouped port model.

terraform {
  required_providers {
    b1fr0st = {
      source  = "RobinHellgren/b1fr0st"
      version = "~> 0.3.0"
    }
  }
}

variable "project_id" {
  type = string
}

provider "b1fr0st" {}

resource "b1fr0st_network_forward" "web" {
  project_id     = var.project_id
  network_name   = "tenant-apps"
  listen_address = "192.0.2.20" # Replace with a reserved address from the Incus uplink.
  listen_port    = 443
  target_address = "10.250.1.20" # Replace with an address assigned to a workload on this network.
  target_port    = 8443
  protocol       = "tcp"
  description    = "HTTPS ingress"
}

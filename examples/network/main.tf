terraform {
  required_providers {
    b1fr0st = {
      source  = "robinhellgren/b1fr0st"
      version = "~> 0.3.0"
    }
  }
}

provider "b1fr0st" {}

variable "project_id" {
  type        = string
  description = "Project UUID from the b1fr0st.dev console."
}

resource "b1fr0st_network" "apps" {
  project_id = var.project_id
  name       = "apps"
}

output "network_name" {
  value = b1fr0st_network.apps.id
}

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

resource "b1fr0st_disk" "data" {
  project_id = var.project_id
  name       = "app-data"
  size_gib   = 40
}

output "disk_id" {
  value = b1fr0st_disk.data.id
}

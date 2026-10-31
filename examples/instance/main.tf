terraform {
  required_providers {
    b1fr0st = {
      source  = "robinhellgren/b1fr0st"
      version = "~> 0.1"
    }
  }
}

provider "b1fr0st" {}

variable "tenant_id" {
  type        = string
  description = "Tenant UUID from the b1fr0st.dev console."
}

resource "b1fr0st_instance" "web" {
  tenant_id = var.tenant_id
  name      = "web"
  type      = "container"
  image     = "ubuntu/24.04"
  profile   = "small"
}

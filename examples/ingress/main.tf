terraform {
  required_providers {
    b1fr0st = {
      source  = "robinhellgren/b1fr0st"
      version = "~> 0.1"
    }
  }
}

provider "b1fr0st" {}

variable "tenant_id" { type = string }

resource "b1fr0st_ingress" "web" {
  tenant_id    = var.tenant_id
  name         = "Web app"
  domain_id    = var.domain_id
  subdomain    = "web"
  site_id      = var.newt_site_id
  instance_name = "web"
  port         = 8080
}

variable "domain_id" { type = string }
variable "newt_site_id" { type = number }

output "hostname" { value = b1fr0st_ingress.web.hostname }
output "certificate" { value = b1fr0st_ingress.web.certificate }

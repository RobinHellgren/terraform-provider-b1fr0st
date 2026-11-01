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
  type = string
}

variable "ssh_public_key" {
  type      = string
  sensitive = true
}

resource "b1fr0st_instance" "worker" {
  project_id = var.project_id
  name      = "worker"
  type      = "virtual-machine"
  image     = "ubuntu/24.04/cloud"
  profile   = "medium"

  cloud_init = "#cloud-config\n${yamlencode({
    packages = ["qemu-guest-agent", "curl"]
    users = [{
      name                = "ubuntu"
      groups              = ["sudo"]
      shell               = "/bin/bash"
      sudo                = "ALL=(ALL) NOPASSWD:ALL"
      ssh_authorized_keys = [var.ssh_public_key]
    }]
    write_files = [{
      path        = "/etc/motd"
      permissions = "0644"
      content     = "Managed by Terraform on b1fr0st.dev\\n"
    }]
    runcmd = [
      ["systemctl", "enable", "--now", "qemu-guest-agent"],
      ["install", "-d", "-m", "0755", "/opt/bootstrap-example"],
    ]
  })}"
}

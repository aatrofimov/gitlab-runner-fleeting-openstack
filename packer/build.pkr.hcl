packer {
  required_plugins {
    openstack = {
      version = ">= 1.1.2"
      source  = "github.com/hashicorp/openstack"
    }
  }
}

source "openstack" "builder" {
  image_name          = var.image_name
  source_image        = var.source_image_id
  flavor              = var.flavor
  ssh_username        = var.ssh_username
  networks            = [var.network_id]
  security_groups     = [var.security_group]
  floating_ip_network = var.external_net
  image_visibility    = "private"
}

build {
  sources = ["source.openstack.builder"]

  provisioner "shell" {
    script = "scripts/install-docker.sh"
  }

  provisioner "shell" {
    script = "scripts/configure-ssh.sh"
  }
}

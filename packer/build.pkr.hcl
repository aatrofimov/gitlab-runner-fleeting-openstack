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
  availability_zone   = var.availability_zone
  ssh_username        = var.ssh_username
  networks            = [var.network_id]
  security_groups     = [var.security_group]
  floating_ip_network = var.external_net
  image_visibility    = "private"

  use_blockstorage_volume = true
  volume_size             = var.volume_size
  volume_type             = var.volume_type

  ssh_keypair_name     = "packer-builder"
  ssh_private_key_file = var.ssh_private_key_file
  user_data_file       = "user-data.yaml"
  config_drive         = true
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

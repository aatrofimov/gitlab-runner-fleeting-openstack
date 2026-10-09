#!/usr/bin/env bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive

sudo apt-get update -y
sudo apt-get install -y docker.io

sudo systemctl enable docker
sudo systemctl start docker

# Let the provisioning user talk to the Docker daemon without sudo.
sudo usermod -aG docker "$(id -un)"

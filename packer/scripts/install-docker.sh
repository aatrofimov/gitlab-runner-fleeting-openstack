#!/usr/bin/env bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive

# Packer on Selectel provisions as root (ssh_username = "root"), and the
# root-centric Debian image does not ship sudo. Use sudo only when we are not
# already root, and fail early with a clear message if it is missing.
if [ "$(id -u)" -eq 0 ]; then
  SUDO=""
elif command -v sudo >/dev/null 2>&1; then
  SUDO="sudo"
else
  echo "error: not running as root and 'sudo' is not installed" >&2
  exit 1
fi

$SUDO apt-get update -y
$SUDO apt-get install -y docker.io

$SUDO systemctl enable docker
$SUDO systemctl start docker

# The runner executes 'docker info' as root over a non-interactive SSH session,
# so the CLI must be present at the standard docker.io location.
if [ ! -x /usr/bin/docker ]; then
  echo "error: docker CLI not found at /usr/bin/docker after install" >&2
  exit 1
fi

# Let the provisioning user talk to the Docker daemon without sudo.
$SUDO usermod -aG docker "$(id -un)"


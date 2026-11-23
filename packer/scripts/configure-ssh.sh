#!/usr/bin/env bash
set -euo pipefail

# Packer on Selectel provisions as root (ssh_username = "root"); sudo may be
# absent from the image. Use it only when we are not already root.
if [ "$(id -u)" -eq 0 ]; then
  SUDO=""
elif command -v sudo >/dev/null 2>&1; then
  SUDO="sudo"
else
  echo "error: not running as root and 'sudo' is not installed" >&2
  exit 1
fi

# Make sure sshd is enabled and running on boot. cloud-init is intentionally
# left intact so the per-instance keypair injection keeps working.
if systemctl list-unit-files ssh.service >/dev/null 2>&1; then
  $SUDO systemctl enable ssh
  $SUDO systemctl start ssh
else
  $SUDO systemctl enable sshd
  $SUDO systemctl start sshd
fi

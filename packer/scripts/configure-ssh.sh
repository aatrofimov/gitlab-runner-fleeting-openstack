#!/usr/bin/env bash
set -euo pipefail

# Make sure sshd is enabled and running on boot. cloud-init is intentionally
# left intact so the per-instance keypair injection keeps working.
if systemctl list-unit-files ssh.service >/dev/null 2>&1; then
  sudo systemctl enable ssh
  sudo systemctl start ssh
else
  sudo systemctl enable sshd
  sudo systemctl start sshd
fi

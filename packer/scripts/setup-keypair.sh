#!/usr/bin/env bash
# Подготовка SSH-ключей для Packer-сборки (идемпотентно).
set -euo pipefail

KEY="${PACKER_BUILDER_KEY:-$HOME/.ssh/packer-builder}"
PUB="$KEY.pub"
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"   # packer/

[ -f "$KEY" ] || ssh-keygen -t ed25519 -f "$KEY" -N "" -C "packer-builder" >/dev/null

if ! openstack keypair show packer-builder >/dev/null 2>&1; then
  openstack keypair create packer-builder --public-key "$PUB" >/dev/null
fi

PUBKEY="$(cat "$PUB")"
cat > "$DIR/user-data.yaml" <<EOF
#cloud-config
ssh_authorized_keys:
  - $PUBKEY
users:
  - name: root
    ssh_authorized_keys:
      - $PUBKEY
EOF

echo "OK: keypair 'packer-builder' + $DIR/user-data.yaml"

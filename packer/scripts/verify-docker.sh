#!/usr/bin/env bash
set -euo pipefail

# Final gate for the image bake: a "dockerless" image must not be built
# unnoticed. These checks mirror what GitLab Runner's docker-autoscaler executor
# does over SSH (ssh root@<ip> 'docker info') plus the CLI location contract.

if [ ! -x /usr/bin/docker ]; then
  echo "error: docker CLI not found at /usr/bin/docker" >&2
  exit 1
fi

docker version
docker info

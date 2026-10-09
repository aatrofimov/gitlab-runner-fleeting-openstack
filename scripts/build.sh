#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

go build -o bin/gitlab-runner-fleeting-openstack .
echo "built bin/gitlab-runner-fleeting-openstack"

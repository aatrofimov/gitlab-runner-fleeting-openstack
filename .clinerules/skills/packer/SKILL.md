---
name: packer
description: Use when creating or modifying Packer HCL2 configurations (openstack builder, provisioners) to bake VM images.
---
# Packer

## Overview
The build-server image bake pipeline (Docker + sshd) lives in `packer/build.pkr.hcl`.

## Conventions
- HCL2, `hashicorp/openstack` builder, auth from `OS_*` / clouds.yaml.
- Provisioning via shell scripts in `packer/scripts/`.
- Stable `image_name` (`gitlab-builder`); rebuilding replaces the image.

## Workflow
1. `packer validate packer/`.
2. `packer build packer/` (requires OpenStack credentials).
3. Update `docs/image-building.md` when the pipeline changes.

## Reference
- Packer openstack builder: developer.hashicorp.com/packer/integrations/hashicorp/openstack

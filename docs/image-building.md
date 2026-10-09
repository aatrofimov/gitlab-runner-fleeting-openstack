# Image building

The build-server image bakes Docker and sshd into a Debian genericcloud image,
so ephemeral builders boot ready. The SSH key is **not** baked in: it is
injected per instance via the OpenStack keypair (cloud-init).

## Prerequisites

- [Packer](https://developer.hashicorp.com/packer/downloads) 1.9+.
- OpenStack credentials in `clouds.yaml` / `OS_*` env.
- A Debian `genericcloud` source image in Glance, plus a flavor, a network, a
  security group (SSH open) and an external network.

## Build

```shell
cd packer
packer init .
packer build \
  -var source_image_id=<debian-genericcloud-uuid> \
  -var flavor=<flavor> \
  -var network_id=<net-uuid> \
  -var security_group=<sg-name> \
  -var external_net=<external-net> \
  .
```

`packer validate .` checks the template without building.

## Update workflow

The baked image keeps the stable name `gitlab-builder`. On rebuild, Packer
replaces the old image with the same name; the plugin resolves the **latest**
active image with that name at startup (via Glance, sorted by `created_at`),
so the next scale-up automatically uses the new image.

## Check the baked image

```shell
openstack image list --name gitlab-builder
openstack image show <image-uuid>
```

Boot a test server from it and confirm Docker is ready:

```shell
ssh debian@<ip> 'docker version'
```

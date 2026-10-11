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

Copy the example variables, fill in real values, then build:

```shell
cd packer
cp example.pkrvars.hcl vars.pkrvars.hcl   # fill in real values
packer init .
packer build -var-file=vars.pkrvars.hcl .
```

`packer validate -var-file=vars.pkrvars.hcl .` checks the template before
building. Note: the openstack plugin authenticates during `validate`, so it
needs real credentials (`clouds.yaml` / `OS_*` env). CI therefore runs only
`packer fmt -check` (syntax), not `validate`.

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

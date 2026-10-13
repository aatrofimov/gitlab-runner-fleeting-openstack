# Image building

The build-server image bakes Docker and sshd into a Debian genericcloud image,
so ephemeral builders boot ready.

The build server's SSH key is injected via **user-data (cloud-config)**,
delivered through the config drive (`config_drive = true`) — **not** via the
OpenStack keypair, which Selectel images ignore (they are oriented toward
password-SSH from the panel). `scripts/setup-keypair.sh` generates the keypair
and writes `user-data.yaml` with the key for both the default user and `root`.

## Prerequisites

- [Packer](https://developer.hashicorp.com/packer/downloads) 1.9+.
- OpenStack credentials in `clouds.yaml` / `OS_*` env.
- A Debian `genericcloud` source image in Glance, plus a flavor, a network, a
  security group (SSH open) and an external network.

## Prepare the SSH key (once)

```shell
bash scripts/setup-keypair.sh
```

Generates `~/.ssh/packer-builder`, creates the `packer-builder` keypair and
writes `packer/user-data.yaml`.

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

## Selectel notes

- Use an **HFL2** flavor with a **local disk** (e.g. `HFL2.2-4096-32-AMD`,
  id `8051`). In `ru-7b` the `standard` (SL) flavors have no available hosts
  (`No valid host was found`), and volume-backed boot is unnecessary — HFL2
  boots from its local disk.
- Set `ssh_username = "root"` in `vars.pkrvars.hcl`: the Selectel Debian image
  is root-centric (`set-root-pw.service`); `debian` is not the default user.

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
ssh root@<ip> 'docker version'
```

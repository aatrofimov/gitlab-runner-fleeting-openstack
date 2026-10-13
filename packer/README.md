# Packer pipeline

Bakes the ephemeral build-server image (Docker + sshd inside). The build
server's SSH key is injected via `user-data.yaml` (cloud-config) through the
config drive, not via the OpenStack keypair.

```shell
bash scripts/setup-keypair.sh                    # generates the key + user-data.yaml (once)
cp example.pkrvars.hcl vars.pkrvars.hcl          # fill in real values
packer init .
packer validate -var-file=vars.pkrvars.hcl .
packer build -var-file=vars.pkrvars.hcl .
```

`packer validate`/`build` need real OpenStack credentials (`clouds.yaml` /
`OS_*`). CI runs `packer fmt -check` for syntax only.

See [docs/image-building.md](../docs/image-building.md) for the full workflow.

# Packer pipeline

Bakes the ephemeral build-server image (Docker + sshd inside, SSH key injected
per instance via keypair).

```shell
cp example.pkrvars.hcl vars.pkrvars.hcl   # fill in real values
packer init .
packer validate -var-file=vars.pkrvars.hcl .
packer build -var-file=vars.pkrvars.hcl .
```

`packer validate`/`build` need real OpenStack credentials (`clouds.yaml` /
`OS_*`). CI runs `packer fmt -check` for syntax only.

See [docs/image-building.md](../docs/image-building.md) for the full workflow.

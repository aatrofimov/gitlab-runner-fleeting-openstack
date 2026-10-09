# Packer pipeline

Bakes the ephemeral build-server image (Docker + sshd inside, SSH key injected
per instance via keypair).

```shell
packer init .
packer validate .
packer build -var-file=vars.pkrvars.hcl .
```

See [docs/image-building.md](../docs/image-building.md) for the full workflow.

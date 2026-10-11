# How it works

This page explains the concepts: what the plugin is and how a job flows
through the system. For the concrete setup steps, see the
[Quick start](../README.md#quick-start).

## The plugin is not a CLI

`gitlab-runner-fleeting-openstack` is a [fleeting](https://gitlab.com/gitlab-org/fleeting/fleeting)
plugin: a gRPC subprocess that GitLab Runner spawns and drives. A human never
runs it directly (apart from `-version`). The runner manager calls the plugin's
`provider.InstanceGroup` methods to create and destroy the ephemeral build
servers.

## The runtime flow of a single job

1. GitLab queues a job.
2. The runner manager fetches it and calls `Increase` to create a VM.
3. The plugin creates a server from the `gitlab-builder` image (flavor +
   networks + security groups), injects the ephemeral ed25519 public key via
   the keypair and tags it `preemptible`.
4. The runner manager calls `ConnectInfo` to get the VM address + private key.
5. The runner manager SSHes to the VM's Docker daemon and runs the job in a
   container there.
6. On completion the runner manager calls `Decrease` to delete the VM. If the
   VM was preempted earlier, the plugin reports it via `Heartbeat`/`Update`
   (`StateDeleted`) and the job is retried on a fresh VM.

## See also

- [Quick start](../README.md#quick-start)
- [Installation](install.md)
- [Configuration](configuration.md)
- [Image building](image-building.md)
- [Development](development.md)

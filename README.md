# gitlab-runner-fleeting-openstack

A [fleeting](https://gitlab.com/gitlab-org/fleeting/fleeting) plugin for
OpenStack that lets [GitLab Runner](https://docs.gitlab.com/runner/)
autoscale ephemeral (preemptible) build servers with the `docker-autoscaler`
executor. Ships with a [Packer](https://www.packer.io/) pipeline that bakes a
ready-made build-server image (Docker + sshd preinstalled) so builders boot
instantly.

- **Generic OpenStack** via [gophercloud](https://github.com/gophercloud/gophercloud);
  Selectel preemptible servers are the primary target.
- Preemptibility = the configurable server tag (`preemptible` by default) with
  Nova microversion `2.72`.
- The SSH key is injected **per instance** through an ephemeral ed25519 keypair
  (cloud-init), never baked into the image.

## Pinned versions

| Component | Version |
|-----------|---------|
| fleeting SDK | `gitlab.com/gitlab-org/fleeting/fleeting v0.0.0-20261007103110-1a9ff2408b83` |
| GitLab Runner | **17.1+** (Docker Autoscaler GA; 16.6+ beta) |
| Go | 1.26+ |

## Build

```shell
scripts/build.sh
# or
CGO_ENABLED=0 go build -o bin/gitlab-runner-fleeting-openstack .
```

## Attach to a runner manager

Use the `docker-autoscaler` executor and point `plugin` at the binary:

```toml
[[runners]]
  executor = "docker-autoscaler"

  [runners.autoscaler]
    plugin = "/usr/local/bin/gitlab-runner-fleeting-openstack"
    capacity_per_instance = 1
    max_instances = 5

    [runners.autoscaler.plugin_config]
      [runners.autoscaler.plugin_config.openstack]
        cloud = "selectel"
        image_name = "gitlab-builder"
        flavor_id = "<flavor-uuid>"
        networks = ["<net-uuid>"]
        security_groups = ["default"]
        assign_floating_ip = true
        floating_ip_network = "<external-net>"
      [runners.autoscaler.plugin_config.instance_group]
        name = "builder"
        max_instances = 5

    [runners.autoscaler.connector_config]
      username = "debian"
      use_static_credentials = false
```

A complete example is in [`examples/config.toml`](examples/config.toml).

## Environment / credentials

Authentication is resolved from `clouds.yaml` and the standard `OS_*`
environment variables (`OS_AUTH_URL`, `OS_USERNAME`, `OS_PASSWORD`,
`OS_PROJECT_NAME`, `OS_REGION_NAME`, ...). See
[`examples/clouds.yaml.example`](examples/clouds.yaml.example).

## Bake the image

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

See [docs/image-building.md](docs/image-building.md).

## Documentation

- [Installation](docs/install.md)
- [Configuration](docs/configuration.md)
- [Image building](docs/image-building.md)
- [Development](docs/development.md)

## Repository

> The Go module path is currently `github.com/trofimovaa/gitlab-runner-fleeting-openstack`;
> update it (and `.goreleaser.yml` ldflags) to the actual repository URL before
> publishing.

## License

MIT.

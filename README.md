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

## Quick start

Step-by-step from a fresh clone to a running autoscaled runner.

### 1. Bake the build-server image

Needs OpenStack credentials (`clouds.yaml` or the `OS_*` env vars). Produces the
`gitlab-builder` Glance image.

```shell
cd packer
packer init .          # once: downloads the openstack plugin
packer build \
  -var source_image_id=<debian-genericcloud-uuid> \
  -var flavor=<flavor> \
  -var network_id=<net-uuid> \
  -var security_group=<sg-name> \
  -var external_net=<external-net> \
  .
```

See [docs/image-building.md](docs/image-building.md).

### 2. Build the plugin and copy it to the runner host

```shell
CGO_ENABLED=0 go build -o gitlab-runner-fleeting-openstack .
scp gitlab-runner-fleeting-openstack user@runner-host:/usr/local/bin/
```

### 3. Register a runner and configure it

On the runner host, install `gitlab-runner` (a separate binary from this
plugin) and register to obtain a token:

```shell
gitlab-runner register --non-interactive \
  --url https://gitlab.example.com \
  --registration-token <PROJECT_OR_GROUP_TOKEN> \
  --executor docker-autoscaler \
  --description "openstack-autoscaler" \
  --docker-image docker:27
```

Then edit `config.toml` and fill in the `[runners.autoscaler]` block — copy it
from [examples/config.toml](examples/config.toml), set the real UUIDs/names and
the path to the plugin binary in `plugin`.

See [docs/install.md](docs/install.md) and
[docs/configuration.md](docs/configuration.md).

### 4. Run the manager

```shell
gitlab-runner run
```

### 5. Test

Queue a job (a project with a `.gitlab-ci.yml` running `docker info`) and watch
a preemptible server appear and disappear:

```shell
openstack server list --name builder
```

Concepts and the end-to-end flow: [docs/how-it-works.md](docs/how-it-works.md).

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

- [How it works](docs/how-it-works.md)
- [Installation](docs/install.md)
- [Configuration](docs/configuration.md)
- [Image building](docs/image-building.md)
- [Development](docs/development.md)

## Repository

> The Go module path is currently `github.com/aatrofimov/gitlab-runner-fleeting-openstack`;
> update it (and `.goreleaser.yml` ldflags) to the actual repository URL before
> publishing.

## License

MIT.

# Configuration

The plugin receives its configuration from GitLab Runner as JSON. In the
runner's `config.toml` you write it as a TOML table under
`[runners.autoscaler.plugin_config]`, and the runner converts it to JSON before
handing it to the plugin. Every key maps to a `json:` tag in
[`config/config.go`](../config/config.go).

## `[runners.autoscaler.plugin_config.openstack]`

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `cloud` | string | `""` | Named cloud in `clouds.yaml`. When empty, the standard `OS_*` environment variables are used. |
| `auth_url` | string | `""` | Optional override of the identity endpoint. |
| `region` | string | `""` | Optional region override. |
| `compute_microversion` | string | `"2.72"` | Nova microversion. `2.72` is required by Selectel to honor the preemptible tag. |
| `image_id` | string | `""` | Glance image UUID to boot. Takes precedence over `image_name`. |
| `image_name` | string | `"gitlab-builder"` | Glance image name; the plugin resolves the latest active image with this name. |
| `flavor_id` | string | — | **Required.** Nova flavor UUID. |
| `keypair_name` | string | `<instance_group.name>-key` | OpenStack keypair that receives the ephemeral public key. |
| `availability_zone` | string | `""` | Optional availability zone. |
| `networks` | []string | `[]` | Neutron network UUIDs to attach to each server. |
| `security_groups` | []string | `[]` | Security group names applied to each server. |
| `assign_floating_ip` | bool | `false` | Allocate and attach a floating IP to each server. |
| `floating_ip_network` | string | `""` | External network (name or ID) for floating IPs. Required when `assign_floating_ip` is true. |
| `preemptible_tag` | string | `"preemptible"` | Nova server tag that marks a server preemptible. |
| `ssh_user` | string | `"debian"` | SSH user fallback when the connector config has no `username`. |
| `max_create_attempts` | int | `5` | Maximum server-create attempts per instance when a server enters the OpenStack `ERROR` state. |
| `create_retry_backoff` | duration string | `"10s"` | Initial delay between create attempts; doubles on each retry, capped at `60s`. |

## `[runners.autoscaler.plugin_config.instance_group]`

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `name` | string | `"builder"` | Instance group name; used as the server name prefix and in the provider ID. |
| `idle_time` | duration string | `""` | Baseline polling/retry interval (informational; the runner drives scaling). |
| `max_instances` | int | `5` | Maximum instances reported as `ProviderInfo.MaxSize`. |

## `[runners.autoscaler.connector_config]`

Connection settings shared by all instances, delivered to the plugin via
`provider.Settings`. The OpenStack plugin uses the ephemeral keypair it
injected at boot, so only `username` is normally required.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `username` | string | `""` | SSH user for the Docker daemon connection. |
| `use_static_credentials` | bool | `false` | Use `password`/`key` instead of the ephemeral keypair. |
| `password` | string | `""` | Password for static credentials. |
| `key` | string | `""` | Private key (PEM) for static credentials. |
| `protocol` | string | `"ssh"` | Connection protocol (`ssh`, `winrm`, `winrm+https`). |
| `protocol_port` | int | `22` | Connection port. |
| `os` / `arch` | string | auto | Instance OS/architecture. |
| `keepalive` / `timeout` | duration | — | Connection keepalive/timeout. |

## Environment variables / clouds.yaml

Authentication is resolved by `gophercloud/utils`'s `clientconfig`, which reads
(in priority order) the named `cloud` entry in `clouds.yaml` and the standard
`OS_*` variables: `OS_AUTH_URL`, `OS_USERNAME`, `OS_PASSWORD`,
`OS_PROJECT_NAME`, `OS_USER_DOMAIN_NAME`, `OS_PROJECT_DOMAIN_NAME`,
`OS_REGION_NAME`, `OS_IDENTITY_API_VERSION`. See
[`examples/clouds.yaml.example`](../examples/clouds.yaml.example).

# Installation

## Prerequisites

- GitLab Runner **17.1+** (Docker Autoscaler executor GA). 16.6+ works in beta.
- An OpenStack cloud with a baked build-server image (see
  [image-building.md](image-building.md)).
- A network, an external network (for floating IPs), a flavor, and a security
  group that allows SSH (22) and the Docker daemon port from the runner
  manager.

## 1. Build the plugin

```shell
scripts/build.sh
# or
CGO_ENABLED=0 go build -o /usr/local/bin/gitlab-runner-fleeting-openstack .
```

## 2. Configure clouds.yaml (or OS_* env)

```shell
mkdir -p ~/.config/openstack
cp examples/clouds.yaml.example ~/.config/openstack/clouds.yaml
$EDITOR ~/.config/openstack/clouds.yaml
```

## 3. Configure the runner manager

Copy [`examples/config.toml`](../examples/config.toml) into the runner's
`config.toml`, fill in the token and cloud specifics, and make sure
`plugin` points at the plugin binary (either an absolute path or a name on
`PATH`).

## 4. Run the runner manager

```shell
gitlab-runner run
```

Each queued job makes the runner scale up a preemptible build server, run the
job against the Docker daemon on that server over SSH, and tear the server
back down.

## Notes

- Preemptible Selectel servers live at most 24 hours and can be reclaimed at
  any moment. Jobs must be idempotent and retryable.
- Give each runner manager (and each `[[runners]]` entry) its own instance
  group name so they never fight over the same servers.

# Development

## Build

```shell
go build ./...
```

## Test

```shell
go test ./...
```

Unit tests cover config validation, the OpenStack wrapper (httptest), and the
plugin lifecycle (fake client). They never require a real cloud.

## Lint / vet / format

```shell
gofmt -w .
go vet ./...
golangci-lint run
```

## Manual end-to-end

1. Bake an image with Packer (see [image-building.md](image-building.md)).
2. Register a runner with the `docker-autoscaler` executor using
   `examples/config.toml`.
3. Run the manager: `gitlab-runner run`.
4. Queue a test job and confirm a preemptible server is created, the job runs,
   and the server is deleted.

Verify with the OpenStack CLI:

```shell
openstack server list --name builder
openstack server show <id>   # check the preemptible tag
```

## Architecture

- `config/` — configuration schema, defaults and validation.
- `internal/openstack/` — thin gophercloud wrapper (auth, servers, keypairs,
  images, floating IPs).
- `internal/plugin/` — `provider.InstanceGroup` implementation (gRPC plugin).
- `packer/` — the build-server image pipeline.

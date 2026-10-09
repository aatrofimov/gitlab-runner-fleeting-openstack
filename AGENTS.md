# AGENTS.md — gitlab-runner-fleeting-openstack

Instructions for AI agents working in this repository.

## Project
A Go plugin for GitLab Runner autoscaling (`docker-autoscaler` executor) plus a
Packer pipeline for the ephemeral build-server image. The plugin is generic
OpenStack (via gophercloud); the target cloud is Selectel (preemptible servers).

## Stack
- Go >= 1.22; gophercloud + gophercloud/utils; the fleeting SDK
  (`gitlab.com/gitlab-org/fleeting/fleeting`).
- Packer (`openstack` builder), logrus, cobra/viper, testify, golangci-lint, GoReleaser.

## Commands
- Build: `go build ./...`
- Test: `go test ./...`
- Lint: `golangci-lint run`
- Format/check: `gofmt -w .` + `go vet ./...`
- Packer: `packer validate packer/`
- Release: `goreleaser release --snapshot --clean`

## Layout
```
main.go
internal/plugin/      # fleeting plugin.Task (gRPC)
internal/openstack/   # gophercloud wrapper
config/               # TOML config + validation
packer/               # build.pkr.hcl + provisioning scripts
examples/             # config.toml, clouds.yaml.example
docs/                 # install/configuration/development/image-building
```

## Code conventions
- `gofmt`, `go vet`, `golangci-lint` must be clean before "done".
- Wrap errors with `fmt.Errorf("...: %w", err)`; never swallow them.
- Pass `context.Context` into all network/long-running calls; explicit timeouts/retries.
- No global state; pass config and dependencies explicitly.
- Packages are isolated: `config` does not depend on `internal/*`.

## Git
- Conventional Commits (`feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `ci:`, `test:`, `build:`).
- `main` is protected; work happens on feature branches and PRs.
- MIT license. No secrets in the repo: OpenStack credentials live only in `OS_*`
  env vars / a local `clouds.yaml`.

## Quality (definition of done)
- `go test ./...`, `golangci-lint run`, `go vet ./...` are clean.
- `packer validate` passes.
- README/docs/examples are up to date and match the code.

## Forbidden
- Committing: binaries, `dist/`, `.env`, `clouds.yaml` with real values, private keys.
- Hardcoding Selectel specifics: the `preemptible` tag and microversion belong in config only.

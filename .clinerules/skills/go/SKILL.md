---
name: go
description: Use when writing, generating, or reviewing Go code — modules, tests, error handling, gRPC, linting — in this project.
---
# Go

## Overview
The project is a Go plugin (fleeting) plus a gophercloud wrapper.

## Conventions
- gofmt / go vet / golangci-lint.
- Wrap errors with `%w`; handle them explicitly.
- context.Context + timeouts.
- Packages: `config`, `internal/openstack`, `internal/plugin`.

## Workflow
1. `go test ./...` before and after changes.
2. `golangci-lint run` and `go vet ./...`.
3. Add a godoc comment for new public API.

## Testing
- testify (assert/require), table-driven tests.
- Mock gophercloud (fake provider / httptest) — no real cloud in unit tests.

## Reference
- gophercloud: github.com/gophercloud/gophercloud (+ gophercloud/utils)
- fleeting SDK: gitlab.com/gitlab-org/fleeting/fleeting

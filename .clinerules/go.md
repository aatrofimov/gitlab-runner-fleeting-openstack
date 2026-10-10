# Go

- Keep `gofmt`, `go vet`, `golangci-lint` clean before "done".
- Wrap errors with `%w`; handle them explicitly; never swallow.
- Pass `context.Context` into all network/long-running calls; explicit timeouts/retries.
- No global state; pass config and dependencies explicitly.
- Tests via `go test ./...` (testify), table-driven; unit tests must not need a real cloud.
- Public API gets godoc comments.
- Pin dependencies in `go.mod`/`go.sum`.

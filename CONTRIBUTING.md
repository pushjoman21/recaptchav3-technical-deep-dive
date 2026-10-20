# Contributing

Read [the analysis](Full_Reverse.md), [tools](docs/TOOLS.md) and
[redactions](docs/REDACTIONS.md) first. Put implementation and tests with their domain.
Keep public examples authored and identify every change to the publication boundary.

Run `go test ./...`, `go vet ./...` and `go run ./cmd/check`.
Format only changed Go files with `gofmt`. The Go checker validates gofmt output,
local document links, forbidden material and the publication manifest.

Do not add captures, original assets, keys, live bindings, decompiler code or private evidence.
Do not describe source recovery as execution coverage or synthetic variants as real
release validation. Unknown semantics must fail explicitly. Avoid broad formatting
or refactoring alongside a functional change.

Run `go run ./cmd/check --write-manifest` only after reviewing every changed file.
Then rerun the checks. It inventories the exact publication bytes except its own file; it is an
integrity aid, not an attestation from an independent auditor.

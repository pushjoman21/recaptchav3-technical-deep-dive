# Offline tools

All commands run from the repository root. The recorded toolchain is Go 1.25.3.
There are no JavaScript runtime or npm dependencies. The optional npm scripts are
aliases for Go checks.

## AST inventory and string reduction

```sh
go run ./cmd/research inventory --input examples/strings.js
go run ./cmd/research deobfuscate --input examples/strings.js --out .work/strings --assume-standard-intrinsics --assume-initialized-decoders --strict
```

Inventory reports function ranges in UTF-8 and UTF-16 coordinates with hashes.
The reducer parses supported string-factory ASTs without executing source. The
intrinsic assumption permits supported standard String behavior. The initialization
assumption permits replacing invariant calls after factory initialization. Neither
flag runs JavaScript.

Mutation, escaping tables, shadowed bindings, unsupported grammar and stateful
closures can prevent rewriting. Reports preserve assumptions and diagnostics.
Strict mode writes its evidence before returning an error on unsupported input.
An existing output directory is rejected. Reports from your own inputs can contain
sensitive data and should not be committed automatically.

## Typed register disassembly

```sh
go run ./cmd/research decode --configuration examples/register/configuration.json --input examples/register/program.bin --out .work/register
```

This creates `program.json`, `constants.json`, `bytecode.bin` and `vm_02_output.asm`
in a new directory. The public profile and opcode assignments are authored fixtures.
Other profile identities fail with `UNSUPPORTED_WIRE_PROFILE`. Production profile
recovery is withheld.

Malformed input retains explicit unknown regions and invalid targets; partial
decoding exits unsuccessfully after writing the diagnostic artifacts. No unknown
opcode is treated as NOP. The distributed command does not emit executable code.
[REDACTED: decompiler and runtime implementation].

## Stack catalogue rendering

```sh
go run ./cmd/research stack --input examples/stack/catalogue.json --out .work/stack
```

This renders two authored handler metadata records. It does not execute stack
bytecode. The listing states `.unresolved CONSUMING_INTERPRETER_AND_PROGRAM`.
The full 77-record research catalogue is separately preserved in the checked-in
[stack listing](../output/vm_01_output.asm), with production locators censored.

## Checks and repository layout

```sh
go test ./...
go vet ./...
go run ./cmd/check
```

The checker validates formatting, local document links/assets, forbidden paths,
sensitive patterns, complete listing structure and `PUBLICATION.json` integrity.
After reviewing changed files, `go run ./cmd/check --write-manifest` refreshes the
manifest. Rerun the checks afterward.

| Path | Responsibility |
| --- | --- |
| `cmd/research/` | Four offline commands and integration tests |
| `cmd/check/` | Publication validation |
| `internal/` | AST, bounded static reduction and output protection |
| `vms/` | Family specifications and retained disassembly code |
| `examples/` | Authored input fixtures; not the captured research programs |
| `output/` | Exactly two full recovered listings, with documented redactions |
| `docs/`, `assets/` | English research, notices and local README artwork |

Original samples, captures, private evidence and operational commands are absent.
See [Redactions](REDACTIONS.md) for the resulting reproducibility limits.

# Validation

The checks below apply to this redacted edition. They do not certify a complete
reverse of every original function, an unseen release or a browser challenge flow.

## Reproducible public checks

| Check | Recorded result | Scope |
| --- | --- | --- |
| `go test ./...` | 34 top-level tests / 60 including subtests; 7 packages; no failures or skips | Static analysis/decoder tests, public integration and full-listing checks |
| `go vet ./...` | Passed | Go packages in this standalone module |
| Full ASM preservation | 11,451 instruction records and 77 handler records retained | Private comparison of mnemonic order/target mapping; public structural checks |
| Publication checks | Passed | Local links/assets, Go formatting, two full redacted ASM files, sensitive patterns and manifest integrity |
| Artwork inspection | Both themes rendered and inspected | Local SVG headers; no external image/font dependencies |

The Go tests include malformed/truncated input, varint behavior, special numeric
values, dynamic fixture opcode assignment, rejected profiles, target checks,
UTF-16 string behavior, complete-listing structure, decompiler exclusion and output
protection. Static string tests include renamed tables, shadowing, mutation,
escaping values, stateful closures and bounded reduction.

Decompiler and executable runtime code are deliberately excluded. An earlier private
publication prototype passed eight authored runtime comparisons, but those tests
are not part of this edition's public suite. No public test executes original or
decompiled JavaScript. No browser or private original sample is needed.

`PUBLICATION.json` inventories every distributed file except itself. Git metadata
and local scratch data are excluded. The checker refuses manifest drift. An
additional private check searched for known captured sample identities and sensitive
markers before the first commit. Neither check is a proof against all inference.

## Reported private evidence

The working study reported 42 register handler bindings, a 279-instruction captured
initializer, two locally generated main streams and later localized live streams.
Original/reconstruction scheduler comparisons covered 21 scenarios. Larger local
main fixtures reached only a subset of their instruction positions.

Those source assets, traces, parameters and production bindings are withheld.
Their counts remain reported findings and are not included in public test totals.
The private working-tree checkpoint had 110 top-level Go tests / 155 including
subtests and 193 JS tests; that larger suite cannot be reproduced from this subset.

## Open requirements

- No original stack consuming dispatcher/program has been recovered.
- The Botguard loader data was absent; no local Botguard implementation is supplied.
- Complete original function-purpose coverage has not been achieved.
- Remaining browser, reentrant scheduler and runtime bank-input paths remain open.
- Native Go extraction/profile parity is incomplete in the working project.
- One release's localization variants do not supply a real independent holdout.
- Production reconstruction is deliberately unavailable in this redacted edition.

Tests that decode successfully establish boundaries for accepted inputs. They do
not establish full runtime path coverage, token validity or a bypass result.

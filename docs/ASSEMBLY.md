# Assembly publication

The publication retains the complete instruction/handler records from the current
working listings. Sensitive operands and production source mappings are censored.
Decompiler code is withheld. There is one ASM file per evidenced family.

## Register listing

[vm_02_output.asm](../output/vm_02_output.asm) contains six distinct streams and one
byte-identical alias, in their recovered instruction order:

| Public stream | Provenance class | Instructions |
| --- | --- | ---: |
| `program_00` | Captured initializer | 279 |
| `program_01` | Controlled local main fixture | 5,272 |
| `program_02` | Controlled local control stream | 1 |
| `program_03` | Controlled local control stream | 1 |
| `program_04` | Second controlled local main fixture | 5,897 |
| `program_05` | Second fixture's control stream | 1 |
| `program_06` | Alias of program_05 | No duplicate instruction block |

Total: 11,451 instruction records. The ten existing `.unknown_string` markers are
retained. They represent missing input context, not publication censorship. A later
private live main observation is discussed in the report but was not part of these
curated input listings; this publication does not silently add an unreviewed stream.

## Notation and censorship

- `00042:` is a local instruction ordinal. Original byte offsets are withheld.
- `L_00042` targets that ordinal in the current program, including end-of-stream labels.
- `R0`, `R1`, etc. are consistent local bank aliases in order of appearance. Production register indexes are withheld.
- `C_000` refers to a constant entry in the same program.
- `"[REDACTED:S0001]"` replaces a sensitive string. Repeated markers preserve equality without publishing the value.
- `?N0001` replaces a sensitive number. Small generic control literals remain visible.
- Original program identities, source hashes, seeds, arities and binding locations are censored.

String values outside a small generic-language/property allowlist are censored,
including encoded pools and provider-specific names. Large numeric literals and
guard/recurrence operands are censored. Sensitive values are still present privately;
their public markers are not hashes that could be mistaken for original values.

The private publication check compared instruction counts and mnemonic order against
the original listings and mapped every target from the original verified boundary
table to its public ordinal. The public checker independently validates stream
counts, labels, constant references and unresolved markers. This validates structural
preservation; it cannot prove behavioral equivalence after operands were censored.

## Stack listing

[vm_01_output.asm](../output/vm_01_output.asm) retains all 77 selected handler records
and their semantic purposes. Real registration IDs, dispatcher helper locators and
source dependencies are censored. `H_000` identifies a catalogue record, not a
bytecode address. The source has not supplied a proven consuming stack interpreter.

The listing retains `.unresolved CONSUMING_INTERPRETER_AND_PROGRAM`. It must not be
presented as a full recovered stack program. Redaction does not explain this research
gap; the missing consumer remains a separate unresolved finding.

## Use and limitations

These are annotated research listings, not executable or reassemblable artifacts.
The retained offline tools and authored fixtures demonstrate the disassembly model
without distributing the production extraction pipeline. See
[Validation](VALIDATION.md) and [Redactions](REDACTIONS.md).

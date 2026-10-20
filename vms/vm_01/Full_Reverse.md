# Stack handler fragments

## Established evidence

The working source inventory selected 77 registrations exposing 121 IDs. The
handlers manipulate a stack, frame-like structures, host calls and encrypted
immediate readers. Independent controlled comparisons covered 38 core operations.
Those observations establish stack-oriented fragments; a complete consuming
interpreter and original bytecode program have not been recovered.

[REDACTED: real registration IDs, selectors, source locations, immediate cipher
parameters and handler extraction]. The two authored CLI fixture entries use invented
IDs and authored source references. Their purposes explain subtraction order and
JavaScript addition coercion without claiming to be an executable stack program.

## What the tool does

[assembly.go](assembly.go) accepts a JSON `HandlerCatalogue`, checks qualification
metadata and renders one block per handler. A block contains a display ID, semantic
name, registration metadata, purpose, source witness and required state components.
It checks empty/multiline fields and refuses a mismatched expected source identity.

Qualification is supplied by the caller. Rendering a catalogue is not an independent
proof that the source extraction is correct. Registration IDs remain metadata;
they do not become instruction opcodes or bytecode addresses.

The output begins with `.unresolved CONSUMING_INTERPRETER_AND_PROGRAM`. Its `H_` labels
identify catalogue entries. The public [research listing](../../output/vm_01_output.asm)
retains all 77 handler entries with censored source mappings. The CLI has a separate
two-record authored fixture; it is not the provenance of the research listing.

## Missing semantics

Without a consuming interpreter, the study cannot establish actual instruction
boundaries, branch addressing, call/return conventions, closure environments,
exception regions or self-modification rules for an original stack program.
Literal null callees in the supplied source are preserved as a finding; replacing
them with a guessed function would alter behavior.

No full stack decompiler is supplied. A source-handler catalogue and independently
tested arithmetic core do not close that requirement. The next research step is
to obtain a source-qualified consuming path or runtime trace that proves byte
consumption, handler selection and owner state together.

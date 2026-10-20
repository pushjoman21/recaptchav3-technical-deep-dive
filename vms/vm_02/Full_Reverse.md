# Register-stream VM

## Architecture and configuration

This family uses a byte cursor and a register bank. The examined runtime initially
allocates 2,048 registers but permits indexed writes beyond that initial size.
Handlers consume typed operands; some require a register reference, others a literal,
and some read the raw operand index rather than its register value.

`Configuration` separates numeric opcode bindings from common semantic layouts.
Each entry carries a semantic name, operand bounds, destination behavior, special
index/register operands and optional canonical operator IR. Handler evidence is
retained as JSON metadata. [REDACTED: production bindings and profile extraction].

The public `SyntheticProfileSHA256` identifies the authored fixture profile. The
validator rejects other identities, duplicate/negative opcodes, unknown semantics,
layout mismatches and invalid operator trees. It does not independently prove a
caller's semantic classification; that obligation belongs to configuration recovery.

## Reader behavior

The implementation in [decoder.go](decoder.go) consumes an operand-count byte and
a signed int32-style varint opcode. A destination-bearing instruction then requires
the destination field marker and register index. Operands are length-delimited
typed atoms. Registers, booleans, strings, variable integers and fixed numeric
representations have distinct tags.

The reader permits up to ten varint groups and accumulates the observed int32
semantics, rather than assuming an unbounded integer. Boolean decoding checks all
groups. Fixed int32 and float64 numbers use little-endian decoding. NaN, infinities
and negative zero receive explicit JSON-safe representations. Invalid UTF-8 fails;
the original nonfatal text decoder's malformed-input behavior remains outside the
accepted subset. Negative register indexes are rejected.

Relative control targets are computed from the end of the entire instruction.
Targets must land on an instruction boundary or the stream end. An operand that
requires a raw index cannot be replaced with an ordinary register read. Nested
handler mapping reuses an encoded destination payload; the IR records this unusual
read boundary instead of flattening it away.

All actual wire tags and checks are explicit in the source and exercised by authored
tests. Public fixture opcode numbers were invented for the example. No release's
opcode assignment table is distributed.

## Instruction representation

| Field | Meaning |
| --- | --- |
| `address`, `end`, `raw` | Half-open byte range and exact encoding |
| `opcode`, `semantic`, `mnemonic` | Configured slot, normalized operation and display name |
| `headerEnd`, `operandsStart` | Observable reader boundaries |
| `destination`, `operands` | Typed register/literal inputs and any destination |
| `flow`, `target` | Fallthrough, branch, closure, scheduled entry or yield behavior |
| `operation` | Canonical expression for mapped operators |
| `evidence` | Source/configuration witnesses supplied by extraction |
| `unknownRegions`, `invalidTargets` | Explicit decoding failures; no NOP substitution |

Potential closure/task entries are hints derived from instructions. They do not
restore original function names. This public native subset records flow edges;
it does not claim a complete structured CFG-to-loop decompiler or native liveness
pass. The larger private reference pipeline contains further analyses.

## Semantic operations

| Operation | Behavior and visible state |
| --- | --- |
| `MOVE`, `MOVE_REGISTER` | Read one operand and write a destination; the latter requires register operands |
| `ADD` | Preserve JavaScript `+`, including string concatenation and coercion |
| `BINARY_MAP_LEFT` | Apply canonical operator IR to a scalar or elements of a mappable left operand |
| `CALL_METHOD` | Evaluate receiver, property and arguments in order; preserve method receiver |
| `GET_PROPERTY`, `SET_PROPERTY` | Read/write a property; getters, proxies and exceptions remain observable |
| `CALLBACK_RESULT` | Send the current result through the callback |
| `CALLBACK_REGISTERS` | Read configured registers in order and send their values |
| `LOAD_NULL` | Store JavaScript null |
| `SET_RESULT` | Replace the current result field |
| `SET_GLOBAL`, `GET_GLOBAL` | Access the explicit scope/global adapters |
| `CALL_GLOBAL` | Call a global with the reconstructed global receiver |
| `BRANCH_EQ` | Read offset and comparands, then use JavaScript loose equality |
| `BRANCH_LT` | Read offset and comparands, then use JavaScript less-than semantics |
| `STRING_HASH` | Iterate UTF-16 code units and preserve signed 32-bit hash arithmetic |
| `CHAR_CODES` | Produce code units from a string |
| `REGEXP` | Construct a regular expression through an explicit constructor adapter |
| `NOT` | Apply JavaScript logical negation |
| `MAP_HANDLER` | Invoke a configured nested handler over collection elements; preserve reused cursor boundaries |
| `COMPACT_FORMAT` | Apply the reconstructed compact recursive formatter |
| `CLOCK`, `ELAPSED` | Use the injected clock and elapsed state |
| `MAKE_CLOSURE` | Capture entry and argument slot; invocation routes through a scheduler adapter |
| `SCHEDULE` | Add pending work for an entry and argument bank slot |
| `YIELD` | Preserve the continuation cursor and leave the current slice |
| `CONSTRUCT` | Evaluate constructor/arguments and preserve JavaScript construction behavior |
| `CONCAT_INDEXED` | Concatenate properties selected by a sequence of keys |
| `LCG_CLOSURE` | Create a Number-arithmetic recurrence closure from explicit operands |
| `CALL_FUNCTION` | Evaluate function/arguments and call with null receiver |
| `GUARDED_VALUE` | Wrap a value with token/state and stack-observation checks |
| `TYPEOF` | Apply JavaScript typeof to the read register value |

Property access, coercion, construction, calls and callbacks can throw. The reader
cursor advances when operands are read, so moving an update past an exception can
change the state. The emitter preserves this sequencing explicitly. Scheduler
exception routing is described in [its report](SCHEDULER.md).

## Assembly

The output is this VM's documented assembly notation, not x86 or ARM. A line carries
a byte address, normalized mnemonic and readable operands. `R<n>` names a register;
`L_<hex>` names a target; long string constants receive `.const` definitions with
their full value. Short strings use JSON quoting. No string is silently truncated.

Aliases reflect proven semantics: `MOVE` becomes `LOAD_CONST` for a literal and
`MOV` for a register. Canonical operators become `ADD`, `XOR` and similar names.
Noncanonical operator trees retain their explicit IR rather than receiving a
misleading alias. Equal finite numeric literals can establish an unconditional
loose-equality branch; identical register operands cannot establish it because NaN
and observable register access remain possible.

String reduction recognizes a constrained sequence over a known bank. It preserves
UTF-16 behavior and reports missing inputs or unpaired output units. A decoded string
annotation is tied to both program and source identity. Missing context remains
explicit. Production string values are censored in the complete recovered listing;
see [Assembly publication](../../docs/ASSEMBLY.md).

## Decompilation

The private decompiler parses an authored runtime template and lowers each
instruction into JavaScript. Literal/register reads and cursor assignments appear
in the emitted body. A switch over the current byte position preserves residual
control flow. Output line mappings identify the instruction that produced each
function. These functions are implementation units, not restored source functions.

The private runtime template contains JavaScript semantics for coercions, closures, property
access and calls. Go controls parsing and emission; it does not execute captured JS.
The emitted program can call host adapters, so it must be treated as executable code.
Decompiler and executable runtime code are [REDACTED: implementation withheld].
Public tests do not execute emitted JavaScript.

Incomplete IR and mismatched handler identities prevent emission. The private emitter
remains low-level: high-level loops, exception regions and original lexical
structure are not fabricated. See [Validation](../../docs/VALIDATION.md) for what the
public and private comparisons actually cover.

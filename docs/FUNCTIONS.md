# Function guide

This guide covers the public tools and authored runtime entry points. The original
bundle's 3,482-function atlas is withheld; only 216 original function purposes had
been reviewed at the working checkpoint. It would be inaccurate to present this
edition as an explanation of every original function.

## Go commands and shared analysis

| Function or group | File | Purpose |
| --- | --- | --- |
| `run` | [CLI](../cmd/research/main.go) | Select one of four offline commands; diagnose unknown/withheld commands |
| `inventory` | [CLI](../cmd/research/main.go) | Parse an input, hash it and list functions with byte/UTF-16 coordinates |
| `deobfuscateSource` | [CLI](../cmd/research/main.go) | Run the static reducer, write a report and preserve strict-mode diagnostics |
| `decode` | [CLI](../cmd/research/main.go) | Validate authored configuration, decode and render ASM/IR without executable output |
| `stack` | [CLI](../cmd/research/main.go) | Render handler metadata and state the absent consuming program |
| `readJSONBytes`, `readJSON` | [CLI](../cmd/research/main.go) | Bound JSON reads and reject trailing values |
| `Read`, `Parse`, `Evidence` | [AST source](../internal/jsast/source.go) | Bound input size, parse without evaluation and map UTF-8/UTF-16 ranges |
| `Functions` and its visitor | [AST source](../internal/jsast/source.go) | Discover named/anonymous function ranges and hashes |
| `LastValue`, `Member`, `Name` | [AST source](../internal/jsast/source.go) | Inspect sequence/member structure without rewriting the source |
| `Bind` and scope helpers | [Binding](../internal/jsast/bind.go) | Resolve declarations and lexical references for analysis |
| `Recover` | [Static recovery](../internal/deobfuscate/recover.go) | Find structural string factories; record assumptions, values and unresolved paths |
| `makeIndex`, visitors, stability/invariant helpers | [Static recovery](../internal/deobfuscate/recover.go) | Reject mutation/escape cases and establish memoized-call conditions |
| `recoverOne`, reducer helpers | [Recovery](../internal/deobfuscate/recover.go), [reducer](../internal/deobfuscate/reducer.go) | Evaluate the accepted AST grammar with bounded fuel and explicit failure |
| JS value conversion helpers | [Values](../internal/deobfuscate/value.go) | Preserve the reducer's supported coercions and UTF-16 representation |
| `Rewrite` | [Static recovery](../internal/deobfuscate/recover.go) | Apply proven non-overlapping edits; retain original evidence ranges |
| `CreateDirectory`, `WriteFile` and path checks | [Artifacts](../internal/artifact/directory.go) | Refuse overwrite and outputs inside input trees, including resolved links |

## Register package

| Function or group | File | Purpose |
| --- | --- | --- |
| `Configuration.validate`, `semanticLayouts`, `validateOperation` | [Decoder](../vms/vm_02/decoder.go) | Check profile, slot uniqueness, layout and operator grammar |
| Reader methods and `decodeInstruction` | [Decoder](../vms/vm_02/decoder.go) | Consume typed bytes, track cursor boundaries and enforce operand modes |
| `Decode` | [Decoder](../vms/vm_02/decoder.go) | Build instruction IR, validate targets and preserve unknown regions |
| `Constants`, `Assembly` and mnemonic helpers | [Assembly](../vms/vm_02/assembly.go) | Render readable operands, labels and full string constants |
| `DecodeStrings`, numeric helpers | [String reduction](../vms/vm_02/strings.go) | Reduce constrained code-unit loader sequences from a known bank |
| `Assembly` | [Stack renderer](../vms/vm_01/assembly.go) | Render a qualified catalogue, not a stack program |

## Authored JavaScript runtime

These entry points describe the private authored runtime. Its code is withheld;
this table explains responsibilities and does not expose original source locators.

| Entry point | Purpose |
| --- | --- |
| `stringHash` | Accumulate a signed 32-bit hash over UTF-16 units |
| `mixHash` | Apply the reconstructed Number/bitwise hash mixing |
| `compactFormat` | Serialize the supported compact value forms recursively |
| `applyOperator` | Evaluate canonical operator IR; preserve short-circuit operators |
| `defaultStackFrames` | Supply an explicit host stack observation when no fixture adapter is provided |
| `GuardedValue` | Store a value behind token/state and stack-observation checks |
| `readRegister` | Read a bank entry and honor guarded-value access behavior |
| `readOperand` | Advance the cursor and distinguish ordinary reads from raw indexes |
| `makeWork` | Build a task record with injectable clock-derived defaults |
| `createState` | Allocate VM state and install explicit globals, clocks and callbacks |
| `executeInstruction` | Implement the normalized instruction semantics |
| `runState` | Run until stream end or instruction budget exhaustion; save continuation |
| `runProgram` | Reject incomplete IR before creating and running state |

## Authored scheduler

These entry points describe the private authored scheduler. Its code and executable
tests are withheld from this edition.

| Entry point | Purpose |
| --- | --- |
| `createSchedulerDomain` | Maintain alternate register banks through `swap` and domain-wide `reset` |
| `addTiming` | Update timing count, total and maximum |
| `createClockWatchdog` | Trigger a reset on first getter read, then replace the accessor |
| `requestWork` | Choose urgent/deferred host dispatch from the current work deadline |
| `enqueueWork` | Promote priorities/deadlines and enqueue a task |
| `flushPending` | Move pending task records into the active queue |
| `runSlice` | Resume tasks with bank swaps, budgets, expiry checks and timing updates |
| `attachScheduler` | Validate host adapters and install closure, timer and loader hooks |

Closures installed by `attachScheduler` bind these operations to one state. The
loader default raises a diagnostic. No retained function fetches a production
program, collects a browser fingerprint or submits a token.

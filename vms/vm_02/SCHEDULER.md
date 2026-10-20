# Scheduler

The private authored scheduler runtime models cooperative work around
the register VM. It uses injected host adapters and decoded programs. Production
loader entry points and their source selectors are [REDACTED: operational locations].

## State and lifecycle

A work item carries arguments, an entry cursor, a destination argument slot, a
priority, deadline, expiry and two callbacks. Pending work waits separately from
the active queue. Enqueueing can promote earlier queue entries so ordering depends
on both priority and deadline. The next item is selected from the queue's end.

`createSchedulerDomain` stores alternate register banks in a WeakMap. `swap` exchanges
a state's visible bank with its alternate bank. `reset` replaces the map. It is a
domain-level operation, so it can affect multiple states sharing that domain.
`createClockWatchdog` installs a getter that replaces itself on its first read and
invokes a reset callback. Later reads use an ordinary writable property.

`runSlice` samples the clock, chooses eligible work, invokes a supplied program loader
if required, checks expiry and computes an instruction budget from timing history.
It swaps banks before execution and afterward. An argument bank slot is initialized
once per work record. Work that yields or exhausts its budget saves its next cursor.
Completed work invokes its completion callback; expired or failed load paths follow
the separately modelled result callback path.

The slice records elapsed time and idle/overtime statistics. A clock moving backward
can trigger watchdog behavior. Controlled fixture clocks are required to compare
these paths; a wall-clock-only test is not deterministic evidence.

## Host boundary

| Adapter | Purpose |
| --- | --- |
| `defer(fn, ms, ...args)` | Register deferred execution; its return value can affect short-circuit behavior |
| `requestIdle(fn, options)` | Supply an idle callback when the mode uses idle scheduling |
| `hidden()` / `microtask(fn)` | Route hidden-document work when supplied |
| `loadProgram(data, state)` | Install a decoded program from an explicit context |
| `execute(state)` | Run the interpreter or emitted equivalent |
| `beginMark` / `endMark` | Optional instrumentation callbacks |

`attachScheduler` validates required adapters. Its default loader throws a concrete
diagnostic. It does not fetch a script or infer missing program context. Closures
created by VM instructions enqueue tasks through this scheduler or an explicitly
provided fixture adapter.

The execution wrapper catches interpreter failures and reports the current result
through the modeled callback path. This behavior must not be confused with a full
recovery of source-level try/catch/finally regions inside an original application.

## Evidence limits

The private study compared 21 scheduler scenarios, including closures and pending
work. Those original traces are withheld. The earlier private publication prototype also exercised bank
isolation, the one-shot watchdog, queue promotion, adapter validation and yield/resume
on authored IR. Those tests and executable runtime were removed from this edition
when the owner requested decompiler/runtime code remain private.

All browser lifecycle variants, reentrant loading paths and arbitrary host callbacks
have not been covered. Runtime source inclusion is not a claim that these gaps are
closed. The [main report](../../Full_Reverse.md) retains the unresolved requirements.

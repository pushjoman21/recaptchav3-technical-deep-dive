# Redactions

This edition was assembled from an explicit source allowlist. Private data was
never copied into the repository and then hidden with `.gitignore`. The internal
working sources, original samples and evidence remain unchanged.

`[REDACTED: category]` marks a withheld boundary. Authored examples are labelled
synthetic; they are not substituted live values. Hashes belonging to public fixtures
and dependencies remain because they identify distributed material.

| Category | Publication treatment | Consequence |
| --- | --- | --- |
| Live keys, seeds, nonces and runtime parameters | [REDACTED: runtime cryptographic context] | No live payload decryption or encryption recipe |
| Original assets and exact release/sample identities | [REDACTED: captured assets and release binding] | Private source findings cannot be reproduced from this tree alone |
| Original byte offsets, selectors, minified entry points and hook locations | [REDACTED: operational source locations] | Semantic explanations replace lookup recipes |
| Production opcode binding tables and their extractor | [REDACTED: production configuration] | The shipped decoder accepts an authored synthetic identity only |
| Full recovered ASM | All instruction/handler records retained; sensitive values, source offsets and bank indexes censored | Structure remains readable; listings are not executable |
| Payloads, headers, cookies, tokens and browser recordings | [REDACTED: live request material] | No replayable session artifacts |
| Raw fingerprints, slot maps and device-specific values | [REDACTED: fingerprint data and bindings] | Only semantic categories and information-loss analysis remain |
| Collector, fetcher, solver and token/proof workflows | [REDACTED: operational tooling] | No browser capture, token generation or challenge submission command |
| Absolute workstation paths, usernames in captures and private evidence links | [REDACTED: local provenance] | Only repository-relative public references are used |
| Private logs, task registers, backups and historical scratch outputs | Excluded from the publication tree | They remain in the internal working directory |

## Retained detail

The public code retains generic AST parsing, static string reduction, typed operand
decoding, IR and assembly rendering. [REDACTED: decompiler and executable runtime code].
Wire tags and semantic operations are technical findings, not credentials. Arbitrary
numeric test operands, synthetic registration IDs and a fixture profile hash remain
visible. They are documented as test data and do not identify a production instance.

The full listings replace production byte offsets and register indexes with local
ordinals/aliases. The source-to-production locator map, program bindings and original
dispatch selectors are withheld. See [Assembly publication](ASSEMBLY.md).

## Checks

`go run ./cmd/check` rejects forbidden paths, absolute workstation paths, credential
shapes, operational collector commands and manifest drift. An additional private
publication check compared the tree against known captured identities and sensitive
markers. These checks supplement the source allowlist and manual review; they cannot
prove the absence of every possible sensitive inference.

The first commit contains only the redacted edition. No unredacted Git history is
imported. Private GitHub visibility is an additional access setting, not a substitute
for these redactions.

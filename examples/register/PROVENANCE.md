# Authored register fixture

This 28-byte program and its opcode assignments were authored for this edition.
They contain no captured instruction stream, key, seed, fingerprint or production
binding. The fixture loads 7 into R3, copies it to R4 and maps addition over the
scalar 4 with right operand 5, producing 9 in R5.

The profile hash identifies the string `synthetic-register-stream-v1`. It is a
fixture identity, not the hash of a captured JavaScript release. The binary hash
in generated IR identifies these exact authored program bytes.

# Reporting sensitive material

Do not attach a live payload, token, cookie, fingerprint or unredacted source locator
to a public issue. Use a private report to the repository owner and describe the
category first. This repository does not supply a production collector or solver.

Publication changes must pass the allowlist, redaction review and manifest checks.
An exposed secret would require removal from accessible history and any appropriate
rotation; hiding it with an ignore rule does not remove existing Git objects.

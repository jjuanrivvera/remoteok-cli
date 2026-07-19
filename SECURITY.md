# Security Policy

## Supported versions
The latest released minor version receives security fixes.

## Reporting a vulnerability
Please report suspected vulnerabilities privately via GitHub Security Advisories
(the "Report a vulnerability" button on the repository's Security tab) rather than a public
issue. You will get an acknowledgement within a few days.

## Token / credential handling
`remoteok` talks to an **unauthenticated public API** — it stores and transmits **no
credentials** of any kind. There is no keyring, no token, and no secret in the config file
(which holds only a base-URL and User-Agent override plus user aliases). The `--dry-run`
curl output therefore contains no secret to redact.

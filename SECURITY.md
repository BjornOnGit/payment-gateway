# Security Policy

We take security seriously and appreciate responsible disclosure.

## Supported Versions

We generally support the latest commit on `main`. If a tagged release process is introduced, we will list supported versions here.

## Reporting a Vulnerability

- Do not open a public issue for security reports.
- Prefer GitHub Security Advisories (private report) if enabled on this repository.
- If advisories are not available, contact the maintainers privately.
  - Option 1: Email the repository owner via the email listed on their GitHub profile
  - Option 2: Open an issue requesting a private contact channel (without disclosing details)

Please include:

- A clear description of the issue and potential impact
- Steps to reproduce
- Any relevant logs, PoCs, or environment details

We aim to:

- Acknowledge receipt within 3–5 business days
- Investigate and provide a remediation plan
- Credit reporters if desired after a fix ships

## Best Practices

- Avoid sharing secrets in issues or PRs
- Use environment variables and secret stores for credentials
- Rotate tokens/keys regularly

# Security

Security fixes target the latest commit on `main`. There are no separately
maintained release branches yet.

Please report suspected vulnerabilities through
[GitHub private vulnerability reporting](https://github.com/nik-makarov/0go/security/advisories/new).
If private reporting is unavailable, open an issue asking for a private contact
channel, without including exploit details or sensitive information.

Include the affected revision, operating system, shell, terminal emulator, and
steps to reproduce. Avoid sharing real directory names or private filesystem
contents when a minimal example will do.

Run `make audit` to check dependency integrity and known reachable vulnerabilities.
CI also runs this check on pull requests, pushes to `main`, and a weekly schedule.
An audit cannot establish that software is free of vulnerabilities.

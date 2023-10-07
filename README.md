# gdev-setup

Easily set up bare metal local dev environments.

## Configuration

### Commands

Shell commands that:
- exit with `0` when they succeed
- exit with anything else when they fail

### Fixes

All fix commands are run in order exactly one time. `gdev-setup` continues regardless of fix exit code.

Shell commands that:
- exit with `0` when they succeed
- exit with `1` when the fix skipped running
- exit with anything else otherwise

# gdev-setup

Easily set up bare metal local dev environments.

## Example

`go run . --workDir $(pwd)/testdata/happy-path`

## Configuration

### Commands

Command results are determined by exit code.
- `success`: exit with `0`
- `failure`: anything else

### Fixes

Fixes will be attempted at most one time. A skip does not count as an attempt.

Fix results are determined by exit code.
- `success`: exit with `0`
- `skip`: exit with `1`
- `failure`: anything else

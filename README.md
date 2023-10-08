# gdev-setup

Easily set up bare metal local dev environments.

## Example

`go run . --workDir $(pwd)/testdata/happy-path`

## Testing

`go test .`

## Configuration

Setup configuration stored in a YAML file at `.gdev/gdev.setup.yaml`.

Examples include:
- [happy-path](./testdata/happy-path/.gdev/gdev.setup.yaml)
- [some-known-issues](./testdata/some-known-issues/.gdev/gdev.setup.yaml)

### Steps

Step result is determined by the command exit code.
- `success`: exit with `0`
- `failure`: anything else

### Fixes

Fixes will be attempted at most one time. A skip does not count as an attempt.

Fix commands run with these environment variables set:
- `STEP_LOG_PATH`: the absolute path to the combined stdout/stderr output

Fix result is determined by the command exit code:
- `success`: exit with `0`
- `skip`: exit with `1`
- `failure`: anything else

### Known Issues

Known issues will be displayed when all fixes have been attempted.

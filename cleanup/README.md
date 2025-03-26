# InterSystems Cleanup Action

This action performs cleanup operations on an InterSystems Cache/IRIS namespace by removing packages, globals, and returning unused space.

## Prerequisites

- InterSystems Cache or IRIS instance must be installed and running
- `csession` or `irissession` must be available in the PATH

## Usage

```yaml
- uses: webmais-sistemas/setup-intersystems/cleanup@v1
  with:
    namespace: 'MYAPP'      # Required: Target namespace to cleanup
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| namespace | Target namespace to cleanup | Yes | - |

## Cleanup Operations

The action performs the following cleanup operations:

1. Deletes the "test" package
2. Removes all globals in the namespace (except cspRule)
3. Returns unused space to the database
4. Dismounts the database

## Example Workflow

```yaml
name: Cleanup
on: [workflow_dispatch]  # Manual trigger

jobs:
  cleanup:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Cleanup Namespace
        uses: webmais-sistemas/setup-intersystems/cleanup@v1
        with:
          namespace: 'MYAPP'
```

## Warning

⚠️ This action performs destructive operations that cannot be undone. Use with caution!

## Error Handling

The action will fail if:
- InterSystems instance is not accessible
- Namespace does not exist
- Cleanup operations fail
- Any other InterSystems errors occur

All errors will be displayed in the GitHub Actions log.

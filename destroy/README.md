# InterSystems Destroy Action

This action completely removes an InterSystems Cache/IRIS namespace and its associated database. This is a destructive operation that cannot be undone.

## Prerequisites

- InterSystems Cache or IRIS instance must be installed and running
- `csession` or `irissession` must be available in the PATH
- User must have %SYS privileges to delete namespaces and databases

## Usage

```yaml
- uses: webmais-sistemas/setup-intersystems/destroy@v1
  with:
    namespace: 'MYAPP'      # Required: Namespace to destroy
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| namespace | Namespace to destroy | Yes | - |

## Destroy Operations

The action performs the following destroy operations:

1. Switches to %SYS namespace
2. Verifies namespace exists
3. Gets the database directory path
4. Deletes the namespace configuration
5. Deletes the database configuration
6. Deletes the database files
7. Removes the database directory

## Example Workflow

```yaml
name: Destroy namespace
on: 
  workflow_dispatch:  # Manual trigger only
  schedule:
    - cron: '0 0 * * 0'  # Weekly destroy

jobs:
  destroy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Destroy namespace
        uses: webmais-sistemas/setup-intersystems/destroy@v1
        with:
          namespace: 'MYAPP'

```

## ⚠️ Warning

This action performs **irreversible** destructive operations:
- The namespace will be completely removed
- All associated database files will be deleted
- All data in the namespace will be lost
- This operation cannot be undone

Use this action with extreme caution, especially in production environments.

## Error Handling

The action will fail if:
- InterSystems instance is not accessible
- User doesn't have sufficient privileges
- Namespace doesn't exist
- Database files are in use
- Any other InterSystems errors occur

All errors will be displayed in the GitHub Actions log.

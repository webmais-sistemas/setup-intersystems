# InterSystems Check Integrity Action

This action checks the integrity of InterSystems Cache/IRIS projects by identifying missing foreign keys in persistent classes.

## Prerequisites

- InterSystems Cache or IRIS instance must be installed and running
- `csession` or `irissession` must be available in the PATH

## Usage

```yaml
- uses: webmais-sistemas/setup-intersystems/check-integrity@v1
  with:
    namespace: 'MYAPP'      # Required: Target namespace for the integrity check
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| namespace | Target namespace for the integrity check | Yes | - |

## Output Format

The action generates a text file containing missing foreign key information in the following format:

```
Classe: [Package/Class.cls]
ForeignKey fk{nome_da_fk}(PropertyName) References ReferencedClass();
```

## Example Workflow

```yaml
name: Check Integrity
on: [push]

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Check Project Integrity
        uses: webmais-sistemas/setup-intersystems/check-integrity@v1
        with:
          namespace: 'MYAPP'
```

## Error Handling

The action will fail if:
- InterSystems instance is not accessible
- Namespace does not exist
- SQL query execution fails
- Any other InterSystems errors occur

All errors will be displayed in the GitHub Actions log.

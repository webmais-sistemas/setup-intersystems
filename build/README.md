# InterSystems Build Action

This action builds InterSystems Cache/IRIS projects using Go. It creates a namespace, imports source files, and sets up the environment for testing.

## Prerequisites

- InterSystems Cache or IRIS instance must be installed and running
- `csession` or `irissession` must be available in the PATH

## Usage

```yaml
- uses: webmais-sistemas/setup-intersystems/build@v1
  with:
    namespace: 'MYAPP'      # Required: Target namespace for the build
    snpm: false              # Optional: Enable SNPM MonitorTools (default: false)
    web: false               # Optional: Enable web server (default: false)
    inc: false               # Optional: Enable includes (default: false)

```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| namespace | Target namespace for the build | Yes | - |
| snpm | Enable SNPM MonitorTools | No | false |
| web | Enable web server | No | false |  
| inc | Enable includes | No | false |

## Directory Structure

Your project should follow this structure:
```
.
├── src/
│   ├── *.cls    # InterSystems class files
│   └── *.inc    # Include files
├── web/         # CSP files directory
└── test/        # Test files directory
```

## Example Workflow

```yaml
name: Build
on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build Project
        uses: webmais-sistemas/setup-intersystems/build@v1
        with:
          namespace: 'MYAPP'
```

## Error Handling

The action will fail if:
- InterSystems instance is not accessible
- Namespace creation fails
- Source file imports fail
- Any other InterSystems errors occur

All errors will be displayed in the GitHub Actions log.

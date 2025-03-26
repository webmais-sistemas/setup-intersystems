# InterSystems Test Action

This action runs tests for InterSystems Cache/IRIS projects and generates JUnit XML reports. It uses the built-in `%UnitTest` framework to execute tests and provides detailed test results.

## Prerequisites

- InterSystems Cache or IRIS instance must be installed and running
- `csession` or `irissession` must be available in the PATH
- Tests must be written using the InterSystems `%UnitTest` framework

## Usage

```yaml
- uses: webmais-sistemas/setup-intersystems/test@v1
  with:
    namespace: 'MYAPP'      # Required: Target namespace for running tests    
    output-path: ''        # Optional: Path for test results (defaults to 'test-results/test-report.xml')
    generate-report: false  # Optional: Generate JUnit XML report (defaults to false)
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| namespace | Target namespace for running tests | Yes | - |
| output-path | Path for test results | No | test-results/test-report.xml |
| generate-report | Generate test report in JUnit XML format | No | false |

## Test Report Format

The action generates a JUnit XML report that includes:
- Test suites and test cases
- Number of assertions
- Test execution time
- Failure details with error messages
- Test class and method names

## Example Workflow

```yaml
name: Test
on: [push]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run Tests
        uses: webmais-sistemas/setup-intersystems/test@v1
        with:
          namespace: 'MYAPP'
          generate-report: true
          output-path: 'test-results/test-report.xml'
```

## Test Directory Structure

Your project should follow this structure:
```
.
├── src/
│   └── *.cls    # Source files
└── test/        # Test files directory
    └── *.cls    # Test classes extending %UnitTest.TestCase
```

## Error Handling

The action will fail if:
- InterSystems instance is not accessible
- Namespace does not exist
- Test execution fails
- Report generation fails (if enabled)
- Any other InterSystems errors occur

All errors will be displayed in the GitHub Actions log.

# API Tests

This directory contains scripts for testing the AIService API endpoints.

## Files

- `test_apis.sh` - Main script for testing structurize and summarize APIs
- `basic_validation_tests.sh` - Script for testing basic validation scenarios
- `general_scenarios_tests.sh` - Script for testing general structurization and summarization scenarios
- `task_related_tests.sh` - Script for testing job/task operations
- `README.md` - This file

## Prerequisites

Before running the tests, ensure you have the following installed:

- `curl` - for making HTTP requests
- `jq` - for pretty-printing JSON responses (optional)

## Usage

### Run all tests

```bash
./test_apis.sh
```

### Test specific endpoints

- Test only structurize API:
  ```bash
  ./test_apis.sh --struct
  ```

- Test only summarize API:
  ```bash
  ./test_apis.sh --summar
  ```

- Test only health endpoint:
  ```bash
  ./test_apis.sh --health
  ```

- Test all APIs (same as running without arguments):
  ```bash
  ./test_apis.sh --all
  ```

### Run specific test categories

- Run basic validation tests:
  ```bash
  ./basic_validation_tests.sh
  ```

- Run general scenarios tests:
  ```bash
  ./general_scenarios_tests.sh --all
  ```

- Run only summarization scenarios:
  ```bash
  ./general_scenarios_tests.sh --sum
  ```

- Run only structurization scenarios:
  ```bash
  ./general_scenarios_tests.sh --struc
  ```

- Run task-related tests:
  ```bash
  ./task_related_tests.sh
  ```

### View help

```bash
./test_apis.sh --help
```

## Configuration

By default, the scripts send requests to `http://localhost:8080`. To test against a different server:

1. Edit the `BASE_URL` variable in any of the test scripts
2. Or temporarily override it:
   ```bash
   BASE_URL="https://your-api-server.com" ./test_apis.sh
   ```

## Expected Behavior

- The scripts will show colored output indicating success or failure
- Response bodies will be pretty-printed if `jq` is available
- The scripts exit with code 0 on success, non-zero on failure

## Test Coverage

### Basic Functionality Tests (`test_apis.sh`)
- Health endpoint test
- Complex summarize request with cat drawing
- Complex structurize request with file hierarchy
- Basic success cases

### Basic Validation Tests (`basic_validation_tests.sh`)
- Empty boardID validation
- Empty userID validation
- Too many elements validation (>1000)

### General Scenarios Tests (`general_scenarios_tests.sh`)
- Simple and complex summarization scenarios
- Simple and complex structurization scenarios
- Realistic project structures and architectures

### Task-Related Tests (`task_related_tests.sh`)
- Job status retrieval
- Job abortion functionality
- Non-existent job handling
- Multiple jobs workflow
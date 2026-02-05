# API Tests

This directory contains scripts for testing the AIService API endpoints.

## Files

- `test_apis.sh` - Main script for testing structurize and summarize APIs
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

### View help

```bash
./test_apis.sh --help
```

## Configuration

By default, the script sends requests to `http://localhost:8080`. To test against a different server:

1. Edit the `BASE_URL` variable in `test_apis.sh`
2. Or temporarily override it:
   ```bash
   BASE_URL="https://your-api-server.com" ./test_apis.sh
   ```

## Expected Behavior

- The script will show colored output indicating success or failure
- Response bodies will be pretty-printed if `jq` is available
- The script exits with code 0 on success, non-zero on failure

## Notes

- The summarize API test includes a complex board with a cat drawing made of various elements
- The structurize API test includes a complex board with multiple elements and a hierarchical file structure
- Both tests use realistic data that follows the validation rules defined in the API handlers
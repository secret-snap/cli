# Secretsnap Smoke Tests

This directory contains comprehensive smoke/integration tests for the Secretsnap CLI that verify end-to-end functionality across all major features.

## Overview

The smoke tests are designed to be as "real" as possible, using actual CLI binaries and testing against real backend servers. They cover:

- **Local Mode**: Encryption/decryption, passphrase mode, file permissions
- **Cloud Mode**: API integration, project management, sharing, audit logs
- **Security**: Secret leakage prevention, file permissions, key management
- **Performance**: Speed benchmarks for various file sizes
- **UX**: Help text, error messages, default behaviors
- **Backward Compatibility**: Version management and compatibility

## Test Structure

### Test Files

- `smoke_test.go` - Main smoke test suite
- `acceptance_test.go` - Existing acceptance tests (reused)
- `scripts/run_smoke_tests.sh` - Test runner script

### Test Categories

1. **TestSmokeLocalMode** - Local encryption/decryption functionality
2. **TestSmokeCloudMode** - Cloud integration and API features
3. **TestSmokeCloudModeRealLicense** - Cloud features with real license key
4. **TestSmokeAPI** - Direct API endpoint testing
5. **TestSmokeAPIRealLicense** - API endpoints with real license key
6. **TestSmokeSecurity** - Security and privacy verification
7. **TestSmokePerformance** - Performance benchmarks
8. **TestSmokeBackwardCompatibility** - Version compatibility
9. **TestSmokeUX** - User experience and error handling

## Running the Tests

### Prerequisites

1. Go 1.22+ installed
2. CLI binary built (`make build`)
3. Optional: API server running on `http://localhost:8080`
4. Optional: Real license key in `smoke-test-license.key` file for real license tests

### Quick Start

```bash
# Run all smoke tests
./scripts/run_smoke_tests.sh

# Run only local mode tests (no API server required)
./scripts/run_smoke_tests.sh --local-only

# Run specific test
./scripts/run_smoke_tests.sh TestSmokeLocalMode

# Use custom API server
./scripts/run_smoke_tests.sh --api-url http://localhost:9000
```

### Manual Execution

```bash
# Run with Go test directly
cd cli
go test -v -run "TestSmoke" ./smoke_test.go ./acceptance_test.go

# Run specific test category
go test -v -run "TestSmokeLocalMode" ./smoke_test.go ./acceptance_test.go

# Skip cloud tests
SKIP_CLOUD_TESTS=1 go test -v -run "TestSmoke" ./smoke_test.go ./acceptance_test.go
```

## Test Coverage

### ✅ Local Mode Tests

- **Init & Key Cache**

  - `secretsnap init` creates `.secretsnap.json` and `~/.secretsnap/keys.json`
  - Idempotent init (no prompts, no errors)
  - Correct file permissions (0600)

- **Bundle/Unbundle**

  - `secretsnap bundle .env` creates encrypted file
  - Bundle is not plaintext (non-ASCII)
  - `secretsnap unbundle` with correct permissions (0600)
  - File content matches original exactly

- **Run Command**

  - `secretsnap run` executes commands with environment variables
  - No temp files left behind
  - Correct environment variable expansion

- **Passphrase Mode**

  - Bundle/unbundle with passphrase
  - Wrong passphrase fails cleanly
  - No partial file creation on failure

- **Overwrite Protection**

  - Refuses to overwrite without `--force`
  - Clear error messages
  - `--out` flag works for both bundle/unbundle

- **Key Export**
  - `secretsnap key export` outputs base64
  - Warning in stderr, key in stdout

### ✅ Cloud Mode Tests

- **Login & Project Management**

  - `secretsnap login --license DEV-KEY` creates token
  - `secretsnap project create` updates config
  - Token file created with correct permissions

- **Push & Pull**

  - `secretsnap bundle --push` returns version
  - `secretsnap pull` retrieves latest version
  - File content matches original
  - Correct file permissions

- **Sharing & Audit**

  - `secretsnap share` adds user with role
  - `secretsnap audit` shows expected events
  - Audit logs contain proper event types

- **Token Management**
  - Corrupted token fails with clear message
  - Prompts for re-login when needed

### ✅ Real License Tests

The `TestSmokeCloudModeRealLicense` and `TestSmokeAPIRealLicense` tests run the same scenarios as above but with a real license key loaded from `smoke-test-license.key` file:

- **License Key Loading**

  - Reads license key from `smoke-test-license.key` file
  - Skips tests if file doesn't exist (logs message)
  - Skips tests if file is empty or unreadable
  - Uses real license for authentication

- **Same Test Coverage**
  - All login, project creation, push/pull scenarios
  - Token expiry and corruption testing
  - API endpoint authentication and project creation
  - Proper error handling and skip logic

### ✅ API Tests

- **Authentication**

  - `POST /auth/login` with valid license returns JWT
  - Invalid license returns 401 with proper error shape

- **Project Management**

  - `POST /projects` creates project, returns ID
  - Creator becomes owner and member

- **Error Handling**
  - 404, 401, 403 return consistent JSON error shapes
  - No stack traces in error responses

### ✅ Security Tests

- **Secret Leakage Prevention**

  - No secrets in stdout/stderr
  - No secrets in API logs
  - `--verbose` doesn't leak sensitive data

- **File Permissions**

  - `.env` outputs have 0600 permissions
  - Key files have 0600 permissions
  - Config files have appropriate permissions

- **Key Loss Scenarios**
  - Missing key cache provides actionable guidance
  - Wrong passphrase fails without partial output
  - Cloud recovery works even with local key loss

### ✅ Performance Tests

- **Small Files (1KB)**

  - Bundle/unbundle < 50ms

- **Large Files (200KB)**

  - Bundle/unbundle < 300ms

- **Network Operations**
  - Push/pull round-trip < 1s on normal connection

### ✅ UX Tests

- **Help & Documentation**

  - `--help` shows flags and examples
  - Command-specific help available
  - Clear, actionable error messages

- **Default Behaviors**
  - Bundle defaults to `./secrets.envsnap`
  - Unbundle defaults to `./.env`
  - Warnings for existing files

### ✅ Backward Compatibility

- **Version Management**
  - Push v1, then v2
  - `pull --version 1` still works
  - `pull` (latest) returns v2

## Environment Variables

| Variable                 | Description              | Default                 |
| ------------------------ | ------------------------ | ----------------------- |
| `DEV_SECRETSNAP_API_URL` | API server URL           | `http://localhost:8080` |
| `SKIP_CLOUD_TESTS`       | Skip cloud-related tests | (unset)                 |
| `SKIP_API_TESTS`         | Skip API endpoint tests  | (unset)                 |

## Test Data

Each test creates a clean temporary environment with:

- Temporary directory (`/tmp/ssmoke-*`)
- Test `.env` file with sample secrets
- Isolated CLI configuration
- Clean state between tests

## Troubleshooting

### Common Issues

1. **CLI binary not found**

   ```bash
   make build
   ```

2. **API server not available**

   ```bash
   # Skip cloud tests
   ./scripts/run_smoke_tests.sh --local-only
   ```

3. **Permission denied**

   ```bash
   chmod +x scripts/run_smoke_tests.sh
   ```

4. **Go modules not found**
   ```bash
   go mod tidy
   ```

### Debug Mode

Run tests with verbose output:

```bash
go test -v -run "TestSmoke" ./smoke_test.go ./acceptance_test.go
```

### Individual Test Debugging

```bash
# Run single test with verbose output
go test -v -run "TestSmokeLocalMode/1_InitAndKeyCache" ./smoke_test.go ./acceptance_test.go
```

## Integration with CI/CD

The smoke tests can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run Smoke Tests
  run: |
    cd cli
    ./scripts/run_smoke_tests.sh --local-only
  env:
    DEV_SECRETSNAP_API_URL: ${{ secrets.API_URL }}
```

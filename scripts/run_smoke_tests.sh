#!/bin/bash

# Secretsnap Smoke Test Runner
# This script runs comprehensive smoke/integration tests for the CLI

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLI_DIR="$(dirname "$SCRIPT_DIR")"
API_URL="${SECRETSNAP_API_URL:-http://localhost:8080}"
SKIP_CLOUD_TESTS="${SKIP_CLOUD_TESTS:-}"
SKIP_API_TESTS="${SKIP_API_TESTS:-}"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    # Check if CLI binary exists
    if [[ ! -f "$CLI_DIR/bin/secretsnap" && ! -f "$CLI_DIR/secretsnap" ]]; then
        log_warning "CLI binary not found, building..."
        cd "$CLI_DIR"
        make build
    else
        log_info "CLI binary found"
    fi
    
    log_success "Prerequisites check passed"
}

check_api_server() {
    log_info "Checking API server availability..."
    
    if [[ -n "$SKIP_CLOUD_TESTS" ]]; then
        log_warning "Skipping API server check (SKIP_CLOUD_TESTS=1)"
        return 0
    fi
    
    # Try to connect to API server
    if curl -s --max-time 5 "$API_URL/healthz" > /dev/null 2>&1; then
        log_success "API server is available at $API_URL"
    else
        log_warning "API server not available at $API_URL"
        log_warning "Cloud tests will be skipped"
        export SKIP_CLOUD_TESTS=1
        export SKIP_API_TESTS=1
    fi
}

run_tests() {
    log_info "Running smoke tests..."
    
    cd "$CLI_DIR"
    
    # Set environment variables for tests
    export SECRETSNAP_API_URL="$API_URL"
    export SKIP_CLOUD_TESTS="$SKIP_CLOUD_TESTS"
    export SKIP_API_TESTS="$SKIP_API_TESTS"
    
    # Run the smoke tests
    if go test -v -run "TestSmoke" ./smoke_test.go ./acceptance_test.go; then
        log_success "All smoke tests passed!"
    else
        log_error "Some smoke tests failed"
        exit 1
    fi
}

run_specific_test() {
    local test_name="$1"
    log_info "Running specific test: $test_name"
    
    cd "$CLI_DIR"
    
    # Set environment variables for tests
    export SECRETSNAP_API_URL="$API_URL"
    export SKIP_CLOUD_TESTS="$SKIP_CLOUD_TESTS"
    export SKIP_API_TESTS="$SKIP_API_TESTS"
    
    # Run the specific test
    if go test -v -run "$test_name" ./smoke_test.go ./acceptance_test.go; then
        log_success "Test $test_name passed!"
    else
        log_error "Test $test_name failed"
        exit 1
    fi
}

show_help() {
    cat << EOF
Secretsnap Smoke Test Runner

Usage: $0 [OPTIONS] [TEST_NAME]

Options:
    -h, --help          Show this help message
    --api-url URL       Set API server URL (default: http://localhost:8080)
    --skip-cloud        Skip cloud-related tests
    --skip-api          Skip API endpoint tests
    --local-only        Run only local mode tests

Examples:
    $0                    # Run all smoke tests
    $0 TestSmokeLocalMode # Run only local mode tests
    $0 --local-only       # Run only local mode tests
    $0 --api-url http://localhost:9000  # Use custom API URL

Environment Variables:
    SECRETSNAP_API_URL   API server URL
    SKIP_CLOUD_TESTS     Skip cloud tests if set to 1
    SKIP_API_TESTS       Skip API tests if set to 1

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        --api-url)
            API_URL="$2"
            shift 2
            ;;
        --skip-cloud)
            SKIP_CLOUD_TESTS=1
            shift
            ;;
        --skip-api)
            SKIP_API_TESTS=1
            shift
            ;;
        --local-only)
            SKIP_CLOUD_TESTS=1
            SKIP_API_TESTS=1
            shift
            ;;
        -*)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
        *)
            TEST_NAME="$1"
            shift
            ;;
    esac
done

# Main execution
main() {
    log_info "Starting Secretsnap Smoke Tests"
    log_info "CLI Directory: $CLI_DIR"
    log_info "API URL: $API_URL"
    
    check_prerequisites
    check_api_server
    
    if [[ -n "${TEST_NAME:-}" ]]; then
        run_specific_test "$TEST_NAME"
    else
        run_tests
    fi
    
    log_success "Smoke test run completed!"
}

# Run main function
main "$@"

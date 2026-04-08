#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# ChargeOps API - Test Runner
# Runs all tests inside Docker containers with zero host dependencies.
# =============================================================================

COMPOSE="docker compose"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()  { echo -e "${YELLOW}[TEST]${NC} $*"; }
pass() { echo -e "${GREEN}[PASS]${NC} $*"; }
fail() { echo -e "${RED}[FAIL]${NC} $*"; }

# =============================================================================
# Cleanup
# =============================================================================
cleanup() {
    log "Cleaning up..."
    $COMPOSE down --remove-orphans --volumes --timeout 5 2>/dev/null || true
}

trap cleanup EXIT

# =============================================================================
# Prerequisites
# =============================================================================
log "Checking prerequisites..."
if ! command -v docker &>/dev/null; then
    fail "Docker is not installed"
    exit 1
fi

if ! docker info &>/dev/null; then
    fail "Docker daemon is not running"
    exit 1
fi

# =============================================================================
# Teardown any previous run
# =============================================================================
log "Tearing down any previous test environment..."
$COMPOSE down --remove-orphans --rmi local --volumes 2>/dev/null || true

# =============================================================================
# Build all images
# =============================================================================
log "Building Docker images..."
$COMPOSE --profile test build --no-cache 2>&1
if [ $? -ne 0 ]; then
    fail "Docker build failed"
    exit 1
fi
pass "Docker images built successfully"

# =============================================================================
# Track results
# =============================================================================
UNIT_RESULT=0
API_RESULT=0

# =============================================================================
# Run Unit Tests (no DB dependency)
# =============================================================================
log "Running unit tests..."
$COMPOSE run --rm --no-deps -T test \
    go test -buildvcs=false ./unit_tests/... -v -count=1 2>&1
UNIT_RESULT=$?

if [ $UNIT_RESULT -eq 0 ]; then
    pass "Unit tests passed"
else
    fail "Unit tests failed"
fi

# =============================================================================
# Run API Tests (needs postgres + app running)
# =============================================================================
log "Running API tests..."
# This starts postgres + app via depends_on, then runs tests
$COMPOSE --profile test run --rm -T test \
    go test -buildvcs=false ./API_tests/... -v -count=1 -timeout 300s 2>&1
API_RESULT=$?

if [ $API_RESULT -eq 0 ]; then
    pass "API tests passed"
else
    fail "API tests failed"
fi

# =============================================================================
# Summary
# =============================================================================
echo ""
echo "==========================================="
echo "            TEST RESULTS SUMMARY"
echo "==========================================="

if [ $UNIT_RESULT -eq 0 ]; then
    pass "Unit Tests:  PASS"
else
    fail "Unit Tests:  FAIL"
fi

if [ $API_RESULT -eq 0 ]; then
    pass "API Tests:   PASS"
else
    fail "API Tests:   FAIL"
fi

echo "==========================================="

if [ $UNIT_RESULT -ne 0 ] || [ $API_RESULT -ne 0 ]; then
    fail "OVERALL: FAIL"
    exit 1
else
    pass "OVERALL: PASS"
    exit 0
fi

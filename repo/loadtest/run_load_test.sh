#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# ChargeOps API Load Test
# Target: 50 RPS for 5 minutes, p99 < 300ms
# =============================================================================

API_URL="${API_BASE_URL:-http://app:8080}"
RATE="${LOAD_TEST_RPS:-50}"
DURATION="${LOAD_TEST_DURATION:-5m}"
P99_THRESHOLD_MS="${LOAD_TEST_P99_THRESHOLD:-300}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()  { echo -e "${YELLOW}[LOAD]${NC} $*"; }
pass() { echo -e "${GREEN}[PASS]${NC} $*"; }
fail() { echo -e "${RED}[FAIL]${NC} $*"; }

# Wait for API to be ready
log "Waiting for API at ${API_URL}..."
RETRIES=30
until curl -sf "${API_URL}/health" > /dev/null 2>&1; do
    RETRIES=$((RETRIES - 1))
    if [ $RETRIES -le 0 ]; then
        fail "API not reachable at ${API_URL}"
        exit 1
    fi
    sleep 1
done
pass "API is ready"

# Register and login a test user for authenticated endpoints
log "Setting up test user..."
TS=$(date +%s%N)
EMAIL="loadtest_${TS}@test.com"
curl -s -X POST "${API_URL}/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"${EMAIL}\",\"password\":\"LoadTest1234!\",\"display_name\":\"Load Tester\"}" > /dev/null 2>&1 || true

LOGIN_RESP=$(curl -s -X POST "${API_URL}/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"${EMAIL}\",\"password\":\"LoadTest1234!\",\"device_id\":\"loadtest-device\"}" 2>/dev/null || true)
TOKEN=$(echo "$LOGIN_RESP" | jq -r '.token // empty' 2>/dev/null || true)

if [ -z "$TOKEN" ]; then
    log "WARNING: Could not obtain auth token, authenticated endpoints will 401"
fi

# Build vegeta targets - mix of public and authenticated business endpoints
log "Generating attack targets..."
TARGETS_FILE="/tmp/targets.txt"

cat > "$TARGETS_FILE" <<EOF
GET ${API_URL}/health

POST ${API_URL}/api/v1/auth/login
Content-Type: application/json
@/tmp/login_body.json

GET ${API_URL}/api/v1/content/carousel

GET ${API_URL}/api/v1/content/campaigns

GET ${API_URL}/api/v1/content/rankings
EOF

# Add authenticated targets if we have a token
if [ -n "$TOKEN" ]; then
cat >> "$TARGETS_FILE" <<EOF

GET ${API_URL}/api/v1/users/me
Authorization: Bearer ${TOKEN}
X-Device-Id: loadtest-device

GET ${API_URL}/api/v1/notifications/inbox
Authorization: Bearer ${TOKEN}
X-Device-Id: loadtest-device

GET ${API_URL}/api/v1/orders
Authorization: Bearer ${TOKEN}
X-Device-Id: loadtest-device
EOF
fi

# Create request bodies
cat > /tmp/login_body.json <<EOF
{"email":"nonexistent@test.com","password":"WrongPassword1!","device_id":"load-device"}
EOF

log "Starting load test: ${RATE} RPS for ${DURATION}..."
log "Acceptance criteria: p99 < ${P99_THRESHOLD_MS}ms"
log "Targets: health, auth/login, content (guest), users/me, inbox, orders (authenticated)"
echo ""

vegeta attack -targets="$TARGETS_FILE" -rate="${RATE}/s" -duration="${DURATION}" | \
    vegeta report -type=text > /tmp/vegeta_report.txt 2>&1

cat /tmp/vegeta_report.txt
echo ""

# Parse p99 from report
P99_LINE=$(grep "99th" /tmp/vegeta_report.txt || true)
if [ -z "$P99_LINE" ]; then
    fail "Could not parse p99 from vegeta report"
    exit 1
fi

# Extract p99 value - vegeta reports in various units
P99_VALUE=$(echo "$P99_LINE" | awk '{print $2}')
log "P99 latency: ${P99_VALUE}"

# Convert to milliseconds for comparison
# vegeta reports as e.g. "1.234ms" or "123.456µs" or "1.234s"
P99_MS=$(echo "$P99_VALUE" | awk '
    /µs$/ { gsub(/µs$/, ""); printf "%.2f", $0/1000; next }
    /ms$/ { gsub(/ms$/, ""); printf "%.2f", $0; next }
    /s$/  { gsub(/s$/, "");  printf "%.2f", $0*1000; next }
    { printf "%.2f", $0 }
')

log "P99 latency in ms: ${P99_MS}"

# Compare against threshold
RESULT=$(awk "BEGIN { print (${P99_MS} < ${P99_THRESHOLD_MS}) ? 1 : 0 }")

echo ""
echo "==========================================="
echo "         LOAD TEST RESULTS"
echo "==========================================="
echo "Rate:       ${RATE} RPS"
echo "Duration:   ${DURATION}"
echo "P99:        ${P99_VALUE} (${P99_MS}ms)"
echo "Threshold:  ${P99_THRESHOLD_MS}ms"
echo "==========================================="

if [ "$RESULT" -eq 1 ]; then
    pass "LOAD TEST: PASS (p99 ${P99_MS}ms < ${P99_THRESHOLD_MS}ms)"
    exit 0
else
    fail "LOAD TEST: FAIL (p99 ${P99_MS}ms >= ${P99_THRESHOLD_MS}ms)"
    exit 1
fi

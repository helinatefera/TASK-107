# ChargeOps Pricing & Warehouse Administration API

A production-ready Go Echo backend API for managing EV charging pricing, operational configuration, and warehouse master data for multi-organization operators. Runs fully offline with PostgreSQL persistence, role-based access control, and local observability.

## Quick Start

```bash
docker compose up
```

This single command builds and starts all services. No local tools, packages, or configuration required beyond Docker.

## Services

| Service    | URL                          | Port |
|------------|------------------------------|------|
| API Server | http://localhost:8080         | 8080 |
| PostgreSQL | localhost:5433               | 5433 (host) / 5432 (container) |

**Database credentials:**
- User: `chargeops`
- Password: `chargeops`
- Database: `chargeops`

## Verify the Application is Running

After `docker compose up`, wait for the app service to become healthy, then:

```bash
# 1. Health check
curl -s http://localhost:8080/health
# Expected: {"status":"ok"}

# 2. Register a user
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"SecurePass123!","display_name":"Admin"}'
# Expected: 201 Created with user JSON

# 3. Login
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"SecurePass123!","device_id":"my-device"}'
# Expected: 200 OK with {"token":"...","expires_at":"..."}

# 4. Use the token for authenticated requests
TOKEN="<token from login response>"
curl -s http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Device-Id: my-device"

# 5. Test error format consistency
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"bad","password":"short"}'
# Expected: {"code":400,"msg":"validation failed: ..."}
```

## Running Tests

All tests run inside Docker containers with zero host dependencies:

```bash
./run_tests.sh
```

This script:
1. Tears down any previous test environment
2. Builds all Docker images
3. Starts PostgreSQL and the API server
4. Waits for services to be healthy
5. Runs unit tests inside the test container (no DB dependency)
6. Runs API integration tests against the live server
7. Prints a PASS/FAIL summary
8. Cleans up all containers and volumes

### Test Structure

- **`unit_tests/`** - Unit tests for core logic (password validation, field masking, error handling, pricing calculations, config loading). 30 test cases.
- **`API_tests/`** - Functional API tests (health check, auth flows, security validation, permission enforcement, input validation). ~27 test cases.
- **`run_tests.sh`** - Orchestrates containerized test execution.

## API Reference

All endpoints return errors in the format:
```json
{"code": 400, "msg": "validation failed: Email failed on required"}
```

No stack traces, internal details, or framework internals are ever exposed.

### Identity APIs

The Authentication (`/auth/*`), Users (`/users/*`), and Permissions (`/users/:id/permissions`) APIs together fulfill the complete Identity requirements. Authentication handles registration, login/logout, session management, and password recovery. Users handles profile management and role assignment. Permissions handles granular permission viewing and overrides.

### Authentication (`/api/v1/auth`)

| Method | Path                      | Auth | Description |
|--------|---------------------------|------|-------------|
| POST   | `/api/v1/auth/register`          | No   | Register new user (password: min 12 chars, 3/4 character classes) |
| POST   | `/api/v1/auth/login`             | No   | Login with email, password, and device_id; returns bearer token |
| POST   | `/api/v1/auth/logout`            | Yes  | Invalidate current session |
| POST   | `/api/v1/auth/refresh`           | Yes  | Reset idle timeout (30min); absolute 7-day cap unchanged |
| POST   | `/api/v1/auth/recover`           | No   | Get one-time recovery token (no email/SMS) |
| POST   | `/api/v1/auth/recover/reset`     | No   | Reset password using recovery token |

**Session rules:** Idle timeout 30 minutes, absolute expiry 7 days. Refresh resets idle only.
**Lockout:** 10 failed attempts in 10 minutes locks account for 15 minutes.

### Users & Permissions (`/api/v1/users`)

| Method | Path                       | Roles | Description |
|--------|----------------------------|-------|-------------|
| GET    | `/api/v1/users`                   | admin | List all users |
| GET    | `/api/v1/users/:id`               | admin, self | Get user details |
| PUT    | `/api/v1/users/:id`               | admin, self | Update user profile |
| PUT    | `/api/v1/users/:id/role`          | admin | Assign role (guest/user/merchant/administrator) |
| DELETE | `/api/v1/users/:id`               | admin | Delete user |
| GET    | `/api/v1/users/:id/permissions`   | admin, self | Get effective permissions |
| PUT    | `/api/v1/users/:id/permissions`   | admin | Update user permissions |

### Organizations (`/api/v1/orgs`)

| Method | Path          | Roles | Description |
|--------|---------------|-------|-------------|
| POST   | `/api/v1/orgs`       | admin | Create org (hierarchy via parent_id, unique org_code) |
| GET    | `/api/v1/orgs`       | admin, merchant(own) | List organizations |
| GET    | `/api/v1/orgs/:id`   | admin, merchant(own) | Get organization |
| PUT    | `/api/v1/orgs/:id`   | admin | Update organization |
| DELETE | `/api/v1/orgs/:id`   | admin | Delete organization |

### Warehouses, Zones, Bins (`/api/v1/warehouses`, `/api/v1/zones`, `/api/v1/bins`)

| Method | Path                          | Roles | Description |
|--------|-------------------------------|-------|-------------|
| POST   | `/api/v1/warehouses`                 | admin, merchant | Create warehouse |
| GET    | `/api/v1/warehouses`                 | admin, merchant(own org) | List warehouses |
| GET    | `/api/v1/warehouses/:id`             | admin, merchant(own) | Get warehouse |
| PUT    | `/api/v1/warehouses/:id`             | admin, merchant(own) | Update warehouse |
| DELETE | `/api/v1/warehouses/:id`             | admin | Delete warehouse |
| POST   | `/api/v1/warehouses/:id/zones`       | admin, merchant(own) | Create zone |
| GET    | `/api/v1/warehouses/:id/zones`       | admin, merchant(own) | List zones |
| POST   | `/api/v1/zones/:id/bins`             | admin, merchant(own) | Create bin (unique bin_code per warehouse) |
| GET    | `/api/v1/zones/:id/bins`             | admin, merchant(own) | List bins |
| PUT    | `/api/v1/bins/:id`                   | admin, merchant(own) | Update bin |
| DELETE | `/api/v1/bins/:id`                   | admin | Delete bin |

### Items & Categories (`/api/v1/items`, `/api/v1/categories`)

| Method | Path               | Roles | Description |
|--------|--------------------|-------|-------------|
| POST   | `/api/v1/categories`      | admin | Create category |
| GET    | `/api/v1/categories`      | authenticated | List categories |
| POST   | `/api/v1/items`           | admin, merchant | Create item (standardized item_name, sku) |
| GET    | `/api/v1/items`           | authenticated | List items |
| GET    | `/api/v1/items/:id`       | authenticated | Get item |
| PUT    | `/api/v1/items/:id`       | admin, merchant | Update item |
| DELETE | `/api/v1/items/:id`       | admin | Delete item |

### Units of Measure (`/api/v1/units`)

| Method | Path                  | Roles | Description |
|--------|-----------------------|-------|-------------|
| POST   | `/api/v1/units`              | admin | Create unit of measure |
| GET    | `/api/v1/units`              | authenticated | List units |
| POST   | `/api/v1/units/conversions`  | admin | Create conversion (decimal(18,6), must be positive) |
| GET    | `/api/v1/units/conversions`  | authenticated | List conversions |

### Suppliers & Carriers (`/api/v1/suppliers`, `/api/v1/carriers`)

| Method | Path               | Roles | Description |
|--------|--------------------|-------|-------------|
| POST   | `/api/v1/suppliers`       | admin, merchant | Create supplier (dedup: normalized name + tax_id) |
| GET    | `/api/v1/suppliers`       | admin, merchant | List suppliers (tax_id/address masked for non-admin) |
| GET    | `/api/v1/suppliers/:id`   | admin, merchant | Get supplier |
| PUT    | `/api/v1/suppliers/:id`   | admin, merchant(own) | Update supplier |
| POST   | `/api/v1/carriers`        | admin, merchant | Create carrier (dedup: normalized name + tax_id) |
| GET    | `/api/v1/carriers`        | admin, merchant | List carriers |
| GET    | `/api/v1/carriers/:id`    | admin, merchant | Get carrier |
| PUT    | `/api/v1/carriers/:id`    | admin, merchant(own) | Update carrier |

### Stations & Devices (`/api/v1/stations`, `/api/v1/devices`)

| Method | Path                       | Roles | Description |
|--------|----------------------------|-------|-------------|
| POST   | `/api/v1/stations`                | admin, merchant | Create station |
| GET    | `/api/v1/stations`                | admin, merchant(own org) | List stations |
| GET    | `/api/v1/stations/:id`            | admin, merchant(own) | Get station |
| PUT    | `/api/v1/stations/:id`            | admin, merchant(own) | Update station |
| DELETE | `/api/v1/stations/:id`            | admin | Delete station |
| POST   | `/api/v1/stations/:id/devices`    | admin, merchant(own) | Create device |
| GET    | `/api/v1/stations/:id/devices`    | admin, merchant(own) | List devices |
| PUT    | `/api/v1/devices/:id`             | admin, merchant(own) | Update device |
| DELETE | `/api/v1/devices/:id`             | admin, merchant(own) | Delete device |

### Pricing (`/api/v1/pricing`)

| Method | Path                                | Roles | Description |
|--------|-------------------------------------|-------|-------------|
| POST   | `/api/v1/pricing/templates`                | admin, merchant | Create template (station XOR device) |
| GET    | `/api/v1/pricing/templates`                | admin, merchant(own org) | List templates |
| GET    | `/api/v1/pricing/templates/:id`            | admin, merchant(own) | Get template |
| POST   | `/api/v1/pricing/templates/:id/versions`   | admin, merchant(own) | Create version (energy_rate + duration_rate + service_fee) |
| GET    | `/api/v1/pricing/templates/:id/versions`   | admin, merchant(own) | List versions |
| GET    | `/api/v1/pricing/versions/:id`             | admin, merchant(own) | Get version detail |
| POST   | `/api/v1/pricing/versions/:id/activate`    | admin, merchant(own) | Activate version (sets effective_at) |
| POST   | `/api/v1/pricing/versions/:id/deactivate`  | admin, merchant(own) | Deactivate version |
| POST   | `/api/v1/pricing/versions/:id/rollback`    | admin, merchant(own) | Rollback (creates new version cloning target) |
| POST   | `/api/v1/pricing/versions/:id/tou-rules`   | admin, merchant(own) | Add TOU rule (non-overlapping per day type) |
| GET    | `/api/v1/pricing/versions/:id/tou-rules`   | admin, merchant(own) | List TOU rules |
| DELETE | `/api/v1/pricing/tou-rules/:id`            | admin, merchant(own) | Delete TOU rule |

**Pricing model:** Templates version with monotonic version numbers. Each version includes energy rate (kWh), duration rate (minutes, ceiling-rounded), and fixed service fee. Sales tax is snapshotted from global config. Device-level pricing overrides station-level.

### Orders (`/api/v1/orders`)

| Method | Path                        | Roles | Description |
|--------|-----------------------------|-------|-------------|
| POST   | `/api/v1/orders`                   | authenticated | Create order (immutable pricing snapshot) |
| GET    | `/api/v1/orders`                   | admin, self | List orders |
| GET    | `/api/v1/orders/:id`               | admin, self | Get order with snapshot detail |
| POST   | `/api/v1/orders/:id/recalculate`   | admin | Recalculate against version active at order start |

### Content Modules (`/api/v1/content`)

All three types are structured records with priority, start/end time, and target audience role.

| Method | Path                     | Roles | Description |
|--------|--------------------------|-------|-------------|
| POST   | `/api/v1/content/carousel`      | admin, merchant | Create carousel slot |
| GET    | `/api/v1/content/carousel`      | any | List active carousel slots (filtered by role + time) |
| PUT    | `/api/v1/content/carousel/:id`  | admin, merchant(own) | Update carousel slot |
| DELETE | `/api/v1/content/carousel/:id`  | admin | Delete carousel slot |
| POST   | `/api/v1/content/campaigns`     | admin, merchant | Create campaign placement |
| GET    | `/api/v1/content/campaigns`     | any | List active campaigns |
| PUT    | `/api/v1/content/campaigns/:id` | admin, merchant(own) | Update campaign |
| DELETE | `/api/v1/content/campaigns/:id` | admin | Delete campaign |
| POST   | `/api/v1/content/rankings`      | admin, merchant | Create hot ranking |
| GET    | `/api/v1/content/rankings`      | any | List active rankings |
| PUT    | `/api/v1/content/rankings/:id`  | admin, merchant(own) | Update ranking |
| DELETE | `/api/v1/content/rankings/:id`  | admin | Delete ranking |

### Notifications (`/api/v1/notifications`)

| Method | Path                             | Roles | Description |
|--------|----------------------------------|-------|-------------|
| GET    | `/api/v1/notifications/inbox`           | authenticated | Offline message center inbox |
| POST   | `/api/v1/notifications/inbox/:id/read`  | authenticated | Mark read (increments "opened" stat) |
| POST   | `/api/v1/notifications/inbox/:id/dismiss` | authenticated | Mark dismissed (increments "dismissed" stat) |
| GET    | `/api/v1/notifications/subscriptions`   | authenticated | Per-template subscription preferences |
| PUT    | `/api/v1/notifications/subscriptions/:id` | authenticated | Opt in/out per template |
| POST   | `/api/v1/notifications/templates`       | admin | Create notification template |
| GET    | `/api/v1/notifications/templates`       | admin | List templates |
| GET    | `/api/v1/notifications/stats`           | admin | Delivery analytics (generated/delivered/opened/dismissed) |

**Notification rules:** Quiet hours 9 PM - 7 AM local time. Max 2 messages/day per user per template. Per-template opt-out. Analytics tracked without push services.

### Admin & Observability (`/api/v1/admin`)

| Method | Path                       | Roles | Description |
|--------|----------------------------|-------|-------------|
| GET    | `/api/v1/admin/audit-logs`        | admin | Full audit trail for admin + merchant changes |
| GET    | `/api/v1/admin/config`            | admin | View global configurations |
| PUT    | `/api/v1/admin/config/:key`       | admin | Update global config (audited) |
| GET    | `/api/v1/admin/metrics`           | admin | View request metrics |
| GET    | `/api/v1/admin/metrics/export`    | admin | Export metrics (see details below) |

### Health (`/health`)

| Method | Path      | Auth | Description |
|--------|-----------|------|-------------|
| GET    | `/health` | No   | Returns `{"status":"ok"}` |

### Metrics Export Details

`GET /api/v1/admin/metrics/export` supports the following:

| Parameter     | Description | Default |
|---------------|-------------|---------|
| `format`      | `csv` or `json` | `csv` |
| `since`       | RFC3339 start timestamp | 24 hours ago |
| `until`       | RFC3339 end timestamp | now |
| `path`        | Filter by request path (e.g., `/api/v1/auth/login`) | all |
| `status_code` | Filter by HTTP status code (e.g., `200`) | all |

**Exported fields:** `id`, `method`, `path`, `status_code`, `latency_ms`, `recorded_at`

**Retention policy:** Metrics older than 30 days are automatically purged by the background cleanup worker.

Example:
```bash
curl -s "http://localhost:8080/api/v1/admin/metrics/export?format=json&since=2026-04-01T00:00:00Z&path=/api/v1/auth/login" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "X-Device-Id: my-device"
```

## Load Testing

Run the performance validation test:

```bash
docker compose --profile loadtest run --rm loadtest
```

This executes a containerized [vegeta](https://github.com/tsenart/vegeta) load test:
- **Rate:** 50 requests per second
- **Duration:** 5 minutes
- **Acceptance criteria:** p99 latency < 300ms
- Prints a PASS/FAIL result based on the threshold

## Architecture

```
cmd/server/main.go          Entry point with route wiring
internal/
  config/                    Environment-based configuration
  db/                        PostgreSQL connection + transaction helper
  middleware/                 Auth, RBAC, audit, rate limit, metrics, logging
  model/                     Database entity structs
  repo/                      Data access layer (sqlx)
  service/                   Business logic
  handler/                   HTTP handlers (thin: parse, validate, call service)
  dto/                       Request/response structs with validation
  worker/                    Background workers (notifications, cleanup)
  masking/                   Field-level masking (tax_id, addresses)
  apperror/                  Typed errors + error handler
migrations/                  11 SQL migration files (auto-applied on startup)
```

## Security

- **Passwords:** bcrypt with cost 12, minimum 12 characters, 3 of 4 character classes
- **Sessions:** Per-device, SHA-256 hashed tokens, 30-min idle / 7-day absolute expiry
- **Account lockout:** 15 minutes after 10 failed attempts in 10-minute window
- **Recovery tokens:** Single-use, 1-hour expiry, returned once via API, stored hashed
- **RBAC:** Four roles (Guest, User, Merchant, Administrator) with granular permissions
- **Field masking:** tax_id (last 4 chars) and all stored address fields (single-line and multi-line) consistently masked for non-admins
- **Error handling:** No stack traces, internal details, or framework internals exposed
- **Audit trails:** Full audit for all admin and merchant configuration changes
- **Input validation:** Enforced consistently across all endpoints

## Data Models

- **Organizations:** Hierarchy via parent_id, unique org_code, timezone
- **Warehouses/Zones/Bins:** Unique bin_code per warehouse
- **Items:** Standardized item_name, sku, category_id
- **Units of Measure:** Conversion factors as decimal(18,6), constrained positive
- **Suppliers/Carriers:** Duplicate check using normalized name + tax_id
- **Pricing:** Monotonic versioned templates, non-overlapping TOU windows, immutable order snapshots

## Non-Functional

- Single-node Docker deployment
- PostgreSQL with sqlx-managed transactions
- Structured JSON logging (zerolog)
- Local metrics table (exportable as CSV)
- Target: p99 < 300ms at 50 RPS on commodity hardware

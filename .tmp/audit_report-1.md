# ChargeOps Static Audit Report (2026-04-07)

## 1. Verdict
- Overall conclusion: **Partial Pass**
- Rationale: The repository is substantial and aligned with the business domain, but there are material security/requirement-fit defects (including tenant/order boundary and audit completeness issues) and several high-risk validation/coverage gaps.

## 2. Scope and Static Verification Boundary
- What was reviewed:
  - Documentation and manifests: `README.md`, `docker-compose.yml`, `Dockerfile*`, `Makefile`, `run_tests.sh`, `loadtest/*`
  - Entry points/routing/middleware: `cmd/server/main.go`, `internal/middleware/*`
  - Core modules: handlers/services/repos/models for auth, users/permissions, org/warehouse, station/device, pricing/order, content, notifications, admin
  - Persistence: all `migrations/*.up.sql`
  - Tests: all `unit_tests/*.go`, `API_tests/*.go`
- What was not reviewed:
  - Runtime behavior under a live environment, actual DB state transitions, concurrency races in production, container/network behavior.
- What was intentionally not executed:
  - No project startup, no Docker, no tests, no load tests, no external services.
- Claims requiring manual verification:
  - p99 latency objective at 50 RPS in realistic traffic.
  - Real-world concurrency behavior (e.g., TOU overlap race windows, version creation contention).
  - Operational behavior of background workers over time.

## 3. Repository / Requirement Mapping Summary
- Prompt core goal mapped: offline EV charging pricing + warehouse administration API with auth/RBAC, org-tenant boundaries, pricing versioning/TOU/order snapshots, content + notification policies, PostgreSQL/sqlx transactions, and local observability.
- Main implementation areas mapped:
  - Auth/session/recovery/lockout: `internal/service/auth_service.go`, `internal/middleware/auth.go`, `migrations/001_users_and_auth.up.sql`
  - RBAC/permissions and route guards: `cmd/server/main.go`, `internal/middleware/rbac.go`, `internal/service/user_service.go`
  - Warehouse/master data: `internal/handler/warehouse_handler.go`, `internal/service/*`, `migrations/002-003,008`
  - Pricing/order lifecycle: `internal/handler/pricing_handler.go`, `internal/service/pricing_service.go`, `migrations/004_stations_and_pricing.up.sql`
  - Content/notification and analytics: `internal/service/content_service.go`, `internal/service/notification_service.go`, `internal/worker/notification_worker.go`, `migrations/005-006`
  - Audit/metrics/logging/export: `internal/middleware/audit.go`, `internal/middleware/metrics.go`, `internal/handler/admin_handler.go`, `migrations/007_audit_and_metrics.up.sql`

## 4. Section-by-section Review

### 4.1 Hard Gates

#### 4.1.1 Documentation and static verifiability
- Conclusion: **Pass**
- Rationale: Startup/test/docs are present and mostly consistent with the codebase; route inventory and architecture are documented in depth.
- Evidence:
  - `README.md:5-357`
  - `cmd/server/main.go:53-262`
  - `docker-compose.yml:1-65`
  - `run_tests.sh:1-123`
- Note:
  - Minor inconsistency: README says “7 SQL migration files” but repository contains 8 up-migrations.
  - Evidence: `README.md:327`, `migrations/001_users_and_auth.up.sql:1`, `migrations/008_supplier_carrier_org.up.sql:1`

#### 4.1.2 Material deviation from Prompt
- Conclusion: **Partial Pass**
- Rationale: Core domain is implemented, but several prompt-critical semantics are weakened (notably order tenant boundary, duplicate-check semantics, and audit completeness).
- Evidence:
  - Domain alignment: `README.md:1-4`, `cmd/server/main.go:106-236`
  - Deviations: issues listed in Section 5.

### 4.2 Delivery Completeness

#### 4.2.1 Core explicit requirements coverage
- Conclusion: **Partial Pass**
- Rationale:
  - Covered: registration/login/logout/refresh/recovery, lockout/session config, pricing templates/versions/TOU/rollback/recalc, warehouses/zones/bins, content modules, notifications and delivery stats, admin metrics/audit endpoints.
  - Gaps/risks: order-domain tenant boundary not enforced; duplicate-check semantics differ from prompt; full audit trail intent not consistently met for some admin/merchant mutations.
- Evidence:
  - Covered flows: `cmd/server/main.go:85-236`, `internal/service/auth_service.go:67-269`, `internal/service/pricing_service.go:61-559`, `internal/worker/notification_worker.go:40-108`
  - Gaps: `internal/handler/order_handler.go:16-34`, `internal/service/pricing_service.go:289-378`, `internal/repo/supplier_repo.go:46-55`, `cmd/server/main.go:100-104`

#### 4.2.2 End-to-end 0→1 deliverable vs partial demo
- Conclusion: **Pass**
- Rationale: Repository has structured modules, migrations, Docker deployment, and test suites; not a single-file demo.
- Evidence:
  - Structure: `README.md:311-328`, repository tree under `internal/`, `migrations/`, `API_tests/`, `unit_tests/`.

### 4.3 Engineering and Architecture Quality

#### 4.3.1 Structure and decomposition
- Conclusion: **Pass**
- Rationale: Clear handler/service/repo/module split; middleware stack and route grouping are coherent.
- Evidence:
  - `README.md:313-327`
  - `cmd/server/main.go:72-236`

#### 4.3.2 Maintainability and extensibility
- Conclusion: **Partial Pass**
- Rationale: Architecture is maintainable overall, but critical policies are not centralized enough (e.g., order tenant checks and some audit capture consistency are handler-dependent), increasing regression risk.
- Evidence:
  - Positive: layered modules across `internal/service/*`, `internal/repo/*`
  - Risk: ownership checks absent in order create path: `internal/handler/order_handler.go:16-34`, `internal/service/pricing_service.go:289-378`

### 4.4 Engineering Details and Professionalism

#### 4.4.1 Error handling, logging, validation, API design
- Conclusion: **Partial Pass**
- Rationale:
  - Positive: consistent app error envelope and middleware-based structured logging/metrics.
  - Negative: key boundary validations are missing in high-risk paths (order time/range/tenant scope; duplicate semantics updates), and some metrics/load validation evidence is superficial.
- Evidence:
  - Error envelope: `internal/apperror/errors.go:40-64`
  - Logging: `internal/middleware/logger.go:24-31`
  - Missing order guards: `internal/dto/pricing_dto.go:31-36`, `internal/service/pricing_service.go:341-346`

#### 4.4.2 Product-like service vs demo
- Conclusion: **Pass**
- Rationale: The project shape resembles a real service with migrations, persistence, middleware, worker loops, and admin/observability endpoints.
- Evidence:
  - `cmd/server/main.go:53-262`
  - `migrations/001_users_and_auth.up.sql:1-115`
  - `internal/worker/dispatcher.go:10-14`

### 4.5 Prompt Understanding and Requirement Fit

#### 4.5.1 Business goal and constraint fit
- Conclusion: **Partial Pass**
- Rationale: Strong fit on broad domain scope, but several explicit/implicit constraints are undercut:
  - tenant/permission consistency across all resource domains,
  - duplicate-check semantics (“normalized name + tax_id”),
  - full audit trails for admin/merchant changes.
- Evidence:
  - Prompt-fit features: `README.md:92-357`, `cmd/server/main.go:106-236`
  - Constraint gaps: Section 5.

### 4.6 Aesthetics (frontend-only)

#### 4.6.1 Visual/interaction quality
- Conclusion: **Not Applicable**
- Rationale: Repository is backend-only API service.
- Evidence: no frontend source tree; server-only entrypoint `cmd/server/main.go:1-262`.

## 5. Issues / Suggestions (Severity-Rated)

### Blocker / High

1) **Severity: High**  
**Title:** Order creation lacks tenant/object boundary checks  
**Conclusion:** Fail  
**Evidence:**
- `internal/handler/order_handler.go:16-34` (creates order using caller user_id only)
- `internal/service/pricing_service.go:293-307` (resolves device/station, but no caller org ownership check)
- `cmd/server/main.go:196-200` (orders group allows merchant/user broadly)
**Impact:** Merchant/user can create pricing snapshots for devices/stations outside their own organization if IDs are known, violating consistent permission checks and tenant isolation expectations.  
**Minimum actionable fix:** Enforce org ownership in `CreateOrderHandler`/service by resolving device→station→org and comparing against caller org for non-admin callers.

2) **Severity: High**  
**Title:** Full audit trail coverage is incomplete for key admin user-management mutations  
**Conclusion:** Fail  
**Evidence:**
- Missing `mw.Audit(database)` on:
  - `users.PUT("/:id/role", ...)` `cmd/server/main.go:100`
  - `users.PUT("/:id/org", ...)` `cmd/server/main.go:101`
  - `users.DELETE("/:id", ...)` `cmd/server/main.go:102`
  - `users.PUT("/:id/permissions", ...)` `cmd/server/main.go:104`
**Impact:** Prompt requires full audit trails for admin/merchant configuration changes; critical identity/authorization changes can occur without audit records.  
**Minimum actionable fix:** Add audit middleware to all admin/merchant configuration-mutating endpoints and ensure old/new values are captured.

3) **Severity: High**  
**Title:** Duplicate-check semantics for suppliers/carriers do not implement “normalized name + tax_id” and updates skip duplicate enforcement  
**Conclusion:** Fail  
**Evidence:**
- Duplicate check uses `OR` instead of pair semantics:
  - `internal/repo/supplier_repo.go:46-55`
  - `internal/repo/supplier_repo.go:90-99`
- Update paths do not re-check duplicates before save:
  - `internal/service/supplier_service.go:54-79`
  - `internal/service/supplier_service.go:129-151`
**Impact:** False-positive blocks for legitimate records and inconsistent enforcement on updates; prompt-specific dedup rule not faithfully implemented.  
**Minimum actionable fix:** Enforce duplicate lookup on `(normalized_name, tax_id)` semantics (including update paths), with explicit behavior for `tax_id IS NULL` documented and tested.

4) **Severity: High**  
**Title:** Order input validation omits critical constraints (time ordering, non-negative quantities)  
**Conclusion:** Fail  
**Evidence:**
- DTO only checks presence: `internal/dto/pricing_dto.go:31-36`
- Cost math proceeds directly and can process invalid durations: `internal/service/pricing_service.go:341-346`
**Impact:** Invalid business inputs (e.g., `end_time <= start_time`, negative energy) can yield nonsensical or negative billing snapshots, undermining pricing correctness.  
**Minimum actionable fix:** Add explicit request validation in handler/service for `start < end`, `energy_kwh >= 0`, and policy-safe bounds before calculation.

### Medium

5) **Severity: Medium**  
**Title:** User permission grant model can produce unnamed granted permissions, weakening permission API semantics  
**Conclusion:** Partial Fail  
**Evidence:**
- Granted override not in role map is emitted with empty `Name`: `internal/service/user_service.go:87-92`
- Middleware indexes permissions by name: `internal/middleware/auth.go:94-100`
**Impact:** Permission APIs can return incomplete data; additional grants beyond role defaults are not reliably represented for name-based checks.  
**Minimum actionable fix:** Load permission names for user overrides via join with `permissions` table and enforce non-empty permission names in effective permission output.

6) **Severity: Medium**  
**Title:** Delivery stats can be inflated by repeated read/dismiss actions on the same message  
**Conclusion:** Partial Fail  
**Evidence:**
- Read/dismiss always increments stats after update:
  - `internal/service/notification_service.go:78-89`
  - `internal/service/notification_service.go:91-102`
- Repo updates are unconditional:
  - `internal/repo/notification_repo.go:143-150`
**Impact:** `opened`/`dismissed` analytics may not represent distinct events accurately.  
**Minimum actionable fix:** Make updates idempotent (`WHERE read = FALSE` / `WHERE dismissed = FALSE`) and increment stats only when state changes.

7) **Severity: Medium**  
**Title:** Static performance evidence is weak for prompt’s “typical queries” objective  
**Conclusion:** Cannot Confirm Statistically  
**Evidence:**
- Load test script ultimately attacks only `/health`: `loadtest/run_load_test.sh:72-80`
**Impact:** p99 claim at 50 RPS for typical business endpoints is not substantiated by provided load script.  
**Minimum actionable fix:** Add authenticated and representative endpoint mix (pricing/order/list operations) with explicit p99 assertions and reproducible profile.

### Low

8) **Severity: Low**  
**Title:** README migration count is out of date  
**Conclusion:** Fail (doc accuracy)  
**Evidence:** `README.md:327` vs `migrations/008_supplier_carrier_org.up.sql:1`  
**Impact:** Minor documentation drift.  
**Minimum actionable fix:** Update README migration count and architecture note.

## 6. Security Review Summary

- **Authentication entry points:** **Pass**
  - Registration/login/logout/refresh/recovery/reset exist and are wired.
  - Evidence: `cmd/server/main.go:85-93`, `internal/service/auth_service.go:67-269`

- **Route-level authorization:** **Partial Pass**
  - Strong role-gating on route groups and admin endpoints.
  - Evidence: `cmd/server/main.go:95-236`
  - Gap: route-level checks do not guarantee order-domain tenant scope (see High issue #1).

- **Object-level authorization:** **Partial Pass**
  - Many handlers enforce org ownership/self checks.
  - Evidence: `internal/handler/station_handler.go:76-83`, `internal/handler/warehouse_handler.go:96-99`, `internal/handler/user_handler.go:18-27`
  - Gap: no object-level tenant check on order creation path (issue #1).

- **Function-level authorization:** **Partial Pass**
  - Permission middleware exists and used on admin/user-management routes.
  - Evidence: `internal/middleware/rbac.go:25-47`, `cmd/server/main.go:100-104`, `230-235`
  - Gap: effective permission naming bug can degrade intended permission semantics (issue #5).

- **Tenant / user isolation:** **Partial Pass**
  - User self-protection and many org-scoped resources enforce isolation.
  - Evidence: `internal/handler/user_handler.go:70-72`, `internal/handler/pricing_handler.go:113-115`, `internal/handler/supplier_handler.go:52-55`
  - Gap: order creation cross-org risk (issue #1).

- **Admin / internal / debug protection:** **Pass**
  - Admin routes are role + permission protected; no obvious unprotected debug routes found.
  - Evidence: `cmd/server/main.go:230-236`

## 7. Tests and Logging Review

- **Unit tests:** **Partial Pass**
  - Present for password/masking/error/config/helpers; some are superficial and not directly exercising production service with DB interactions.
  - Evidence: `unit_tests/password_test.go:11-81`, `unit_tests/masking_test.go:9-135`, `unit_tests/pricing_logic_test.go:10-190`

- **API / integration tests:** **Partial Pass**
  - Broad endpoint presence and role checks covered, but high-risk business/security paths are incompletely tested.
  - Evidence: `API_tests/auth_test.go:8-324`, `API_tests/tenant_data_isolation_test.go:8-114`, `API_tests/admin_crud_test.go:10-99`

- **Logging categories / observability:** **Pass**
  - Structured request logging + metrics table + export endpoints + cleanup retention worker.
  - Evidence: `internal/middleware/logger.go:24-31`, `internal/middleware/metrics.go:21-107`, `internal/handler/admin_handler.go:126-216`, `internal/worker/cleanup_worker.go:38-49`

- **Sensitive-data leakage risk in logs / responses:** **Partial Pass**
  - API responses avoid stack traces/internal errors.
  - Evidence: `internal/apperror/errors.go:40-64`
  - Residual risk: generic internal errors are logged verbatim (`log.Error().Err(err)`), so operational log sensitivity depends on upstream error content.
  - Evidence: `internal/apperror/errors.go:62`, `internal/worker/notification_worker.go:44-45`

## 8. Test Coverage Assessment (Static Audit)

### 8.1 Test Overview
- Unit tests exist: `unit_tests/*.go` (Go test).
- API/integration tests exist: `API_tests/*.go` (Go test via HTTP requests).
- Test framework: Go `testing` package.
- Test entry points:
  - `run_tests.sh:69-87`
  - `Dockerfile.test:14`
- Documentation provides test command:
  - `README.md:59-76`

### 8.2 Coverage Mapping Table

| Requirement / Risk Point | Mapped Test Case(s) | Key Assertion / Fixture / Mock | Coverage Assessment | Gap | Minimum Test Addition |
|---|---|---|---|---|---|
| Password policy (>=12, 3/4 classes) | `unit_tests/password_test.go:11-81`, `API_tests/auth_test.go:21-57` | Expected app errors for short/weak passwords | basically covered | No API-level edge tests for Unicode/special-class ambiguity | Add API tests for boundary class combos and 12-char exact invalid variants |
| Auth 401/invalid token | `API_tests/auth_security_test.go:10-30` | 401 assertions on protected endpoints | sufficient | None major | Add matrix across several protected route groups |
| Session refresh/device mismatch semantics | `API_tests/auth_test.go:208-222` | Refresh success path only | insufficient | No mismatch header/body/device, idle/abs expiry tests | Add tests for wrong `X-Device-Id`, expired idle, expired absolute session |
| Account lockout (10 in 10m, 15m lock) | none found | n/a | missing | Critical auth control untested | Add integration tests that perform failed logins and verify lockout window/duration |
| RBAC route authorization (403) | `API_tests/station_auth_test.go:8-61`, `API_tests/admin_test.go:8-93`, `API_tests/pricing_test.go:8-45` | Role-based forbidden assertions | basically covered | Coverage is mostly route-level, not deep object-level | Add object-ownership 403 tests per sensitive domain |
| Object-level user isolation | `API_tests/tenant_isolation_test.go:46-115` | User A cannot access/update User B | sufficient | None major | Add delete/self and permission override edge cases |
| Org tenant isolation (stations/warehouses) | `API_tests/tenant_data_isolation_test.go:8-114` | Merchant A/B access checks | basically covered | No equivalent for pricing/order/supplier/content | Add cross-org access tests for pricing templates, orders, suppliers, content updates |
| Pricing lifecycle (version/TOU overlap/activate/rollback) | `API_tests/pricing_crud_test.go:8-127` | 400 overlap, activate/deactivate/rollback checks | basically covered | No test for recalculation at original order start timestamp | Add end-to-end order snapshot + recalc against historical effective version |
| Order domain authorization + validation | `API_tests/order_auth_test.go:8-49` | Weak assertions (not 401/403, 404 on non-existent) | insufficient | Does not test tenant boundary, invalid times, negative energy | Add robust order creation/ownership/validation tests |
| Notification policy (quiet hours, max 2/day, opt-out suppression) | `API_tests/notification_crud_test.go:8-66`, `notification_auth_test.go:8-92` | Template/subscription access only | insufficient | Core worker policy rules not tested | Add worker-policy integration tests using seeded jobs and deterministic time |
| Sensitive masking in API responses | unit-level masking tests only `unit_tests/masking_test.go:9-135` | function-level assertions | insufficient | No API assertions that non-admin responses mask tax_id/address | Add integration tests for org/supplier/warehouse masking by role |
| Admin audit trail generation | none specific found | n/a | insufficient | No tests verifying audit rows after critical mutations | Add API+DB assertions for audited endpoints incl. user role/org/permissions changes |

### 8.3 Security Coverage Audit
- **Authentication:** **Basically covered** for happy-path + basic unauthorized checks, but lockout/expiry/device-edge behavior is largely untested.
- **Route authorization:** **Basically covered** for many 403 scenarios.
- **Object-level authorization:** **Insufficient**; targeted checks exist for users/stations/warehouses, but pricing/order/supplier/content boundaries are under-tested.
- **Tenant / data isolation:** **Insufficient**; no tests for cross-org order creation path (identified High issue).
- **Admin / internal protection:** **Basically covered** by non-admin 403 tests for admin endpoints.

### 8.4 Final Coverage Judgment
- **Partial Pass**
- Covered well enough: basic auth, many route-level RBAC checks, selected tenant checks.
- Major uncovered risks: lockout/session expiry/device mismatch edge cases, order tenant isolation and validation, notification policy enforcement, audit trail completeness, and masking verification at API level.

## 9. Final Notes
- The codebase is substantial and generally product-shaped, but high-severity defects remain in requirement-critical control planes (tenant boundary, auditing completeness, and dedup/validation semantics).
- Findings above are strictly static and evidence-based; runtime/performance assertions remain manual-verification items.

# Recheck Report: Prior Issues from `Audit_report_1.md`

Date: 2026-04-07  
Mode: Static-only verification (no runtime execution, no tests run, no Docker)

## Verdict
Overall recheck result: **7 fixed, 1 fixed with residual manual verification note**.

## Scope
- Reviewed only previously reported issues in `.tmp/Audit_report_1.md` section 5 (`.tmp/Audit_report_1.md:128-214`).
- Verified current repository code and docs by static inspection.
- Did not execute server/tests/load scripts.

## Issue-by-Issue Status

1. **[Previously High] Order creation lacked tenant/object boundary checks**  
   - **Current status:** **Fixed**  
   - **Evidence:**
     - Handler now passes caller org/admin context: `internal/handler/order_handler.go:35-40`
     - Service now enforces tenant guard for non-admin and checks station org match: `internal/service/pricing_service.go:299-310`
   - **Notes:** Runtime behavior still requires manual verification, but static implementation now includes org boundary checks.

2. **[Previously High] Missing audit middleware on key user-management mutations**  
   - **Current status:** **Fixed**  
   - **Evidence:** `mw.Audit(database)` now present on:
     - `PUT /users/:id/role`: `cmd/server/main.go:100`
     - `PUT /users/:id/org`: `cmd/server/main.go:101`
     - `DELETE /users/:id`: `cmd/server/main.go:102`
     - `PUT /users/:id/permissions`: `cmd/server/main.go:104`

3. **[Previously High] Supplier/carrier duplicate-check semantics and update enforcement**  
   - **Current status:** **Fixed**  
   - **Evidence:**
     - Duplicate checks now use strict `(normalized_name AND tax_id)` pair when `tax_id` present:
       - Suppliers: `internal/repo/supplier_repo.go:49-51`
       - Carriers: `internal/repo/supplier_repo.go:94-96`
     - Update flows now re-check duplicates excluding current row:
       - Suppliers: `internal/service/supplier_service.go:74-83`
       - Carriers: `internal/service/supplier_service.go:157-166`

4. **[Previously High] Order validation omitted time/quantity constraints**  
   - **Current status:** **Fixed**  
   - **Evidence:**
     - Enforced `start_time < end_time`: `internal/handler/order_handler.go:27-29`
     - Enforced `energy_kwh` positive: `internal/handler/order_handler.go:30-31`

5. **[Previously Medium] User permission grants could be unnamed**  
   - **Current status:** **Fixed**  
   - **Evidence:**
     - Permission name lookup from full permissions list and assignment for user-only grants:
       - `internal/service/user_service.go:82-87`
       - `internal/service/user_service.go:93-97`

6. **[Previously Medium] Notification opened/dismissed stats double-count risk**  
   - **Current status:** **Fixed**  
   - **Evidence:**
     - Read path now idempotent guard: `internal/service/notification_service.go:84-87`
     - Dismiss path now idempotent guard: `internal/service/notification_service.go:102-105`

7. **[Previously Medium] Load-test script only hit `/health`**  
   - **Current status:** **Fixed**  
   - **Evidence:**
     - Target file now includes mixed endpoints (health, auth, content, and authenticated business endpoints):
       - Target generation: `loadtest/run_load_test.sh:53-69`
       - Authenticated endpoints (`/users/me`, `/notifications/inbox`, `/orders`): `loadtest/run_load_test.sh:72-87`
       - Vegeta uses generated targets file: `loadtest/run_load_test.sh:99-100`

8. **[Previously Low] README migration count out of date**  
   - **Current status:** **Fixed**  
   - **Evidence:** README now states 8 migrations: `README.md:327`

## Consolidated Conclusion
- All eight previously listed issues now have static evidence of remediation in code/docs.
- One general boundary remains: runtime correctness and performance characteristics still require manual execution to fully verify.

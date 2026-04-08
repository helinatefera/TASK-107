# ChargeOps Pricing & Warehouse Administration API Design

## 1. System Overview

* **Platform**: Offline EV charging pricing, warehouse administration, and operational configuration platform for multi-organization operators.
* **Primary roles**: Guest, User, Merchant, Administrator.
* **Core capabilities**:

  * Authentication and identity management
  * Role-based access control with organization scoping
  * Pricing template and version management for stations/devices
  * Order pricing recalculation against historical active price versions
  * Warehouse master-data management (warehouses, zones, bins, items, units, suppliers, carriers)
  * Offline message center and content-module operations management
  * Delivery analytics without push infrastructure
  * Full audit trails, local metrics, and structured logs

ChargeOps is designed to run fully offline on a single node using Docker. The stack consists of a Go Echo backend and PostgreSQL database, with all observability and background processing handled locally. No external providers are required for authentication recovery, notifications, analytics, or monitoring.

## 2. Design Goals

* Fully functional in an offline environment.
* Secure authentication, authorization, and session lifecycle management.
* Accurate and auditable EV pricing with historical version traceability.
* Strong organization-level isolation for merchant-managed resources.
* Clean separation between API, domain logic, persistence, and background jobs.
* Reliable warehouse and master-data administration with validation and duplicate prevention.
* Local observability through structured logs and metrics export files.
* Maintainable APIs with consistent validation, permission checks, and audit logging.

## 3. High Level Architecture

* **Two-tier service architecture**:

  ```
  API Clients / Admin Console
           │
           ▼
    Go + Echo REST API
           │
           ▼
       PostgreSQL
  ```
* **Additional components**:

  * In-process job scheduler
  * Local metrics tables and export jobs
  * Structured logging to disk/stdout
  * Audit log pipeline
  * Local file export module for metrics/log extracts
* **Offline principle**: No external network calls, no email/SMS providers, no cloud queue, and no external monitoring collector.

## 4. API Architecture

* **API design**: RESTful, versioned under `/api/v1/...`.
* **Resource domains**:

  * `/auth` – registration, login, logout, refresh, password recovery
  * `/users`, `/roles`, `/permissions` – identity and RBAC management
  * `/organizations` – org hierarchy and administration
  * `/warehouses`, `/zones`, `/bins` – warehouse structures
  * `/items`, `/categories`, `/uom` – item dictionary and units
  * `/suppliers`, `/carriers` – master data with duplicate checks
  * `/pricing/templates`, `/pricing/versions`, `/pricing/tou-rules` – EV pricing strategy
  * `/orders` – order pricing snapshots and recalculation
  * `/content-modules` – carousel slots, campaigns, hot rankings
  * `/messages`, `/subscriptions` – offline message center and preferences
  * `/analytics` – delivery effectiveness and local reporting
  * `/audit` – privileged audit access
  * `/metrics/exports` – local exportable metrics files

## 5. Backend Architecture (Go + Echo)

* **Middleware stack**:

  * Trace ID injection
  * Authentication
  * RBAC/permission enforcement
  * Request validation
  * Session timeout enforcement
  * Structured request logging
  * Panic recovery and unified error handling
* **Service layer**:

  * `AuthService` – registration, login, logout, refresh, recovery token reset, lockout
  * `SessionService` – per-device sessions, idle timeout, absolute expiration
  * `UserService` – user CRUD, role assignment, organization membership
  * `OrganizationService` – organization hierarchy and global administration
  * `WarehouseService` – warehouses, zones, bins
  * `ItemService` – item dictionary, category binding, SKU validation
  * `UOMService` – units of measure and conversion factors
  * `PartnerService` – suppliers and carriers duplicate checks
  * `PricingService` – template creation, versioning, activation, deactivation, rollback, pricing resolution
  * `OrderService` – price snapshot generation and recalculation
  * `ContentModuleService` – operational content scheduling and priority ordering
  * `MessageCenterService` – offline inbox, quiet hours, opt-out, frequency caps
  * `AnalyticsService` – generated, delivered, opened, dismissed counters
  * `AuditService` – immutable change records
  * `ExportService` – file generation for local metrics exports
* **Background jobs (in-process)**:

  * Quiet-hours message release worker
  * Daily message-cap reset/check worker
  * Metrics rollup/export worker
  * Expired session cleanup worker
  * Recovery token cleanup worker

## 6. Database Design (PostgreSQL)

* **Core tables**:

  * `organizations` – id, org_code, name, parent_id, timezone, status
  * `users` – id, organization_id, username/email, password_hash, role, status, locked_until
  * `permissions`, `user_permissions`, `roles` – RBAC model
  * `device_sessions` – id, user_id, device_id, refresh_token_hash, created_at, last_activity_at, expires_at, revoked_at
  * `password_recovery_tokens` – id, user_id, token_hash, expires_at, used_at
  * `warehouses` – id, organization_id, code, name
  * `warehouse_zones` – id, warehouse_id, zone_code, name
  * `warehouse_bins` – id, warehouse_id, zone_id, bin_code, status
  * `item_categories` – id, name, status
  * `items` – id, sku, item_name, normalized_item_name, category_id, status
  * `uom_groups` – id, name, base_unit_id
  * `uoms` – id, group_id, code, name, conversion_factor
  * `suppliers` – id, organization_id, name, normalized_name, tax_id_masked, address_masked, status
  * `carriers` – id, organization_id, name, normalized_name, tax_id_masked, address_masked, status
  * `pricing_templates` – id, organization_id, scope_type, station_id, device_id, name, status
  * `pricing_template_versions` – id, template_id, version_number, effective_start_at, effective_end_at, is_active, rollback_source_version_id
  * `pricing_components` – id, version_id, energy_rate, duration_rate, fixed_fee, sales_tax_enabled
  * `tou_rules` – id, version_id, day_type, start_minute, end_minute, rate_modifier
  * `orders` – id, organization_id, station_id, device_id, start_time, end_time, applied_version_id, pricing_snapshot_json, total_amount
  * `content_modules` – id, module_type, priority, target_role, start_at, end_at, payload_json, status
  * `message_templates` – id, code, name, status
  * `message_subscriptions` – id, user_id, template_id, enabled
  * `message_inbox` – id, user_id, template_id, content_json, generated_at, delivered_at, opened_at, dismissed_at, state
  * `delivery_analytics_daily` – id, template_id, local_date, generated_count, delivered_count, opened_count, dismissed_count
  * `audit_logs` – id, actor_user_id, actor_role, action, resource_type, resource_id, before_json, after_json, trace_id, created_at
  * `metrics_exports` – id, export_type, file_path, created_at, created_by
* **Constraints**:

  * Unique `org_code`
  * Unique normalized `bin_code` within warehouse
  * Unique normalized `sku`
  * Positive `conversion_factor`
  * Monotonic `version_number` per pricing template
  * Non-overlapping TOU windows per day type
  * Historical orders reference immutable pricing version and snapshot

## 7. Security Design

* **Authentication & password security**:

  * Password minimum length 12
  * Must include at least 3 of 4 classes: lowercase, uppercase, digit, special
  * Passwords hashed with Argon2id using per-password salt
* **Session security**:

  * Per-device session records
  * Device identity provided by client as stable `device_id`
  * Idle timeout: 30 minutes
  * Absolute expiration: 7 days
  * Refresh resets idle timeout but does not extend absolute lifetime
* **Account protection**:

  * 10 failed attempts within 10 minutes triggers 15-minute lockout
  * Recovery tokens are locally generated, single-use, time-limited, and stored as hashes
* **Authorization**:

  * Guest: public identity flows only
  * User: standard self-scope access
  * Merchant: only own organization’s stations/devices/pricing templates and permitted operational data (scoped by `organization_id`)
  * Administrator: all organizations, users, and global configurations
* **Sensitive-field masking**:

  * `tax_id` and address lines masked in non-admin responses
  * Only administrators can view unmasked values
  * Audit logs store redacted versions of sensitive fields

## 8. Domain Model Details

* **Organization**:

  * Supports hierarchy via `parent_id`
  * `org_code` is globally unique
  * Timezone used for local quiet-hours and TOU interpretation
* **Warehouse**:

  * Belongs to an organization
  * Contains zones and bins
  * Bin codes unique per warehouse after normalization
* **Item Dictionary**:

  * Standardized `item_name`
  * Globally unique SKU
  * Category reference required
* **Unit of Measure**:

  * Decimal conversion factors stored as `decimal(18,6)`
  * Must be positive
  * Based on per-group base unit model
* **Supplier/Carrier**:

  * Duplicate detection by normalized name + tax ID where present
  * Sensitive values masked in ordinary read models
* **Pricing Template**:

  * Scoped to station or device
  * Device-level pricing overrides station-level pricing
  * Versioned with start timestamps and rollback lineage
* **Order**:

  * Stores immutable pricing snapshot with component rates, tax data, TOU rule context, and computed totals

## 9. Pricing and Billing Design

* **Pricing components**:

  * Energy charge (kWh)
  * Duration charge (minutes, rounded up using ceiling rule)
  * Fixed service fee (applied once per order)
* **Time-of-use rules**:

  * Configured per day type
  * Non-overlapping within day type
  * Overnight windows normalized into split ranges if needed
  * Timezone derived from organization (or station if defined)
* **Version lifecycle**:

  1. Create draft version
  2. Validate TOU windows and component data
  3. Activate with effective start timestamp
  4. Deactivate or supersede with later version
  5. Rollback by cloning prior version into a new higher version number
* **Pricing resolution**:

  1. Identify device or station scope
  2. Device-level template takes precedence over station-level
  3. Select version active at order start time
  4. Calculate energy, duration, fixed fee, and tax
  5. Tax rate resolved from local configuration tables
  6. Store applied version reference and immutable snapshot
* **Recalculation rule**:

  * Orders are always recalculated against the version active at the order’s start time

## 10. Operations Configuration and Message Center

* **Content modules**:

  * Types: carousel slots, campaign placements, hot rankings
  * Fields: priority, active window, target audience role, structured payload
  * Deterministic tie-breaker: priority, then start time, then record ID
* **Offline message center**:

  * Messages generated from local templates
  * Stored in recipient inbox for retrieval
  * Delivery defined as persisted and visible in user inbox
* **Subscription policies**:

  * Quiet hours: 9:00 PM–7:00 AM local time (organization timezone)
  * Max 2 messages/day per user per template
  * Per-template opt-out enforced before message generation
* **Interaction analytics**:

  * `generated`
  * `delivered_to_inbox`
  * `opened` (counted once per message)
  * `dismissed` (counted once per message)

## 11. Audit, Logging, and Observability

* **Audit trails**:

  * Record all admin and merchant configuration changes
  * Include actor, role, target resource, before/after state, device/session context, timestamp, and trace ID
  * Sensitive fields remain masked in audit payloads
* **Structured logs**:

  * JSON logs written locally
  * Fields: timestamp, level, trace_id, user_id, org_id, endpoint, latency_ms, error_code
* **Local metrics**:

  * Stored in PostgreSQL rollup tables
  * Exportable as CSV and JSON files
* **Trace IDs**:

  * Generated per request and propagated across service layers and audit entries

## 12. Validation and Error Handling

* **Validation rules**:

  * Consistent request validation across all domains
  * Domain-specific checks for duplicate masters, overlapping TOU windows, invalid effective timestamps, and permission scope
* **Unified error envelope**:

  ```json
  {
    "error": {
      "code": "ERR_CODE",
      "message": "Human readable message",
      "details": {}
    }
  }
  ```
* **Typical error codes**:

  * `ERR_INVALID_CREDENTIALS`
  * `ERR_ACCOUNT_LOCKED`
  * `ERR_SESSION_EXPIRED`
  * `ERR_PERMISSION_DENIED`
  * `ERR_DUPLICATE_BIN_CODE`
  * `ERR_DUPLICATE_MASTER_RECORD`
  * `ERR_TOU_WINDOW_OVERLAP`
  * `ERR_INVALID_PRICE_VERSION`

## 13. Background Jobs Design

* **Scheduler**: In-process tickers coordinated with PostgreSQL state tables.
* **Jobs**:

  * Session expiry and cleanup
  * Recovery token cleanup
  * Quiet-hours deferred delivery release
  * Daily delivery analytics rollups
  * Metrics export generation
* **Concurrency control**:

  * Use transactional claiming with `FOR UPDATE SKIP LOCKED` where appropriate
* **Offline guarantee**:

  * No dependency on external cron, queue brokers, email gateways, or telemetry collectors

## 14. Deployment Design

* **Deployment target**: Single-node Docker deployment.
* **Containers**:

  * `api` – Go Echo application
  * `postgres` – PostgreSQL database
* **Runtime characteristics**:

  * Commodity hardware target
  * 99th percentile latency under 300 ms for indexed typical queries at 50 RPS
  * Local persistent volumes for database, logs, and exports
* **Offline readiness**:

  * No external DNS, cloud storage, or third-party services
  * All required binaries and dependencies packaged into images before deployment

## 15. Testing Strategy

* **Unit tests**:

  * Password policy validation
  * session timeout and lockout logic
  * duplicate-check normalization
  * TOU overlap validation
  * pricing-version selection and rollback
* **Integration tests**:

  * Auth and RBAC flows
  * merchant organization scoping
  * order pricing snapshot creation and recalculation
  * message quiet-hours and opt-out enforcement
  * audit log generation
* **Performance tests**:

  * Typical list/read APIs at 50 RPS
  * pricing lookup and order recalculation hot paths
* **Failure scenarios**:

  * session expiry
  * recovery token reuse
  * conflicting pricing windows
  * duplicate master creation
  * deferred message release after quiet hours

## 16. Future Extensibility

* Multi-language admin console
* Richer pricing dimensions such as reservation or parking penalties
* Optional WebSocket-based local inbox updates
* Archived export bundles for audits and compliance reviews
* Horizontal scale-out with shared PostgreSQL and distributed cache if offline constraints evolve

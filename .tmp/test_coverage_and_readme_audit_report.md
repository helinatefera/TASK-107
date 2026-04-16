# Unified Test Coverage + README Audit (Strict Mode)

## Path Notice
Requested output path: `/.tmp/test_coverage_and_readme_audit_report.md`.
Root path is not writable in this environment, so this report is saved at:
- `/Users/macbookpro/Projects/eaglepoint/Task-107/.tmp/test_coverage_and_readme_audit_report.md`
- `/tmp/test_coverage_and_readme_audit_report.md`

---

## 1) Test Coverage Audit

### Project Type Detection
- README top does not explicitly declare one of: backend/fullstack/web/android/ios/desktop.
- Light static inference:
  - Go Echo backend route wiring in `repo/cmd/server/main.go`
  - No frontend source/test framework files found in repo.
- **Inferred project type: backend**.

### Backend Endpoint Inventory
Source of truth: `repo/cmd/server/main.go`.

Total endpoints: **102**

Domain counts:
- Health: 1
- Auth: 6
- Users: 9
- Orgs: 5
- Warehouses/Zones/Bins: 11
- Categories/Items/Units: 11
- Suppliers/Carriers: 8
- Stations/Devices: 9
- Pricing: 12
- Orders: 4
- Content: 12
- Notifications: 9
- Admin: 5

### API Test Classification
1. **True No-Mock HTTP**
- API tests use `doRequest(...)` over `http.NewRequest` + `http.DefaultClient.Do` in `repo/API_tests/helpers_test.go`.

2. **HTTP with Mocking**
- None found in API test suite.

3. **Non-HTTP (unit/integration without HTTP transport)**
- Unit tests use `httptest` in `repo/unit_tests/apperror_test.go` and `repo/unit_tests/auth_helpers_test.go`.

### Mock Detection Rules Check
- Scanned for `jest.mock|vi.mock|sinon.stub|gomock|mockery|dependency override`.
- No API test mocking detected.
- `httptest` usage is isolated to unit tests.

### API Test Mapping Table

#### Health (1)
| Endpoint | Covered | Test Type | Test Files | Evidence |
|---|---|---|---|---|
| `GET /health` | yes | true no-mock HTTP | `API_tests/health_test.go` | `TestHealthEndpoint` |

#### Auth (6)
| Endpoint | Covered | Test Type | Test Files | Evidence |
|---|---|---|---|---|
| `POST /api/v1/auth/register` | yes | true no-mock HTTP | `API_tests/auth_test.go` | `TestRegister_ValidData` |
| `POST /api/v1/auth/login` | yes | true no-mock HTTP | `API_tests/auth_test.go` | `TestLogin_ValidCredentials` |
| `POST /api/v1/auth/logout` | yes | true no-mock HTTP | `API_tests/auth_test.go` | `TestLogout_ValidToken` |
| `POST /api/v1/auth/refresh` | yes | true no-mock HTTP | `API_tests/auth_test.go` | `TestRefresh_ValidToken` |
| `POST /api/v1/auth/recover` | yes | true no-mock HTTP | `API_tests/auth_test.go` | `TestRecover_ValidEmail` |
| `POST /api/v1/auth/recover/reset` | yes | true no-mock HTTP | `API_tests/auth_test.go` | `TestRecoverReset_ValidTokenAndPassword` |

#### Users (9)
| Endpoint | Covered | Test Type | Test Files | Evidence |
|---|---|---|---|---|
| `GET /api/v1/users` | yes | true no-mock HTTP | `API_tests/user_crud_test.go` | `TestAdminListUsers` |
| `GET /api/v1/users/me` | yes | true no-mock HTTP | `API_tests/auth_session_test.go` | `TestDeviceMismatch_Rejected` |
| `GET /api/v1/users/:id` | yes | true no-mock HTTP | `API_tests/tenant_isolation_test.go` | `TestUserCannotAccessOtherUserProfile` |
| `PUT /api/v1/users/:id` | yes | true no-mock HTTP | `API_tests/tenant_isolation_test.go` | `TestUserCannotUpdateOtherUserProfile` |
| `PUT /api/v1/users/:id/role` | yes | true no-mock HTTP | `API_tests/helpers_test.go` | `promoteUser` |
| `PUT /api/v1/users/:id/org` | yes | true no-mock HTTP | `API_tests/helpers_test.go` | `setUserOrg` |
| `DELETE /api/v1/users/:id` | yes | true no-mock HTTP | `API_tests/user_crud_test.go` | `TestAdminDeleteUser` |
| `GET /api/v1/users/:id/permissions` | yes | true no-mock HTTP | `API_tests/user_crud_test.go` | `TestAdminUpdatePermissions` |
| `PUT /api/v1/users/:id/permissions` | yes | true no-mock HTTP | `API_tests/user_crud_test.go` | `TestAdminUpdatePermissions` |

#### Orgs (5)
| Endpoint | Covered | Test Type | Test Files | Evidence |
|---|---|---|---|---|
| `POST /api/v1/orgs` | yes | true no-mock HTTP | `API_tests/org_crud_test.go` | `TestOrgCRUDLifecycle` |
| `GET /api/v1/orgs` | yes | true no-mock HTTP | `API_tests/org_crud_test.go` | `TestOrgCRUDLifecycle` |
| `GET /api/v1/orgs/:id` | yes | true no-mock HTTP | `API_tests/org_crud_test.go` | `TestOrgCRUDLifecycle` |
| `PUT /api/v1/orgs/:id` | yes | true no-mock HTTP | `API_tests/org_crud_test.go` | `TestOrgCRUDLifecycle` |
| `DELETE /api/v1/orgs/:id` | yes | true no-mock HTTP | `API_tests/org_crud_test.go` | `TestOrgCRUDLifecycle` |

#### Warehouses/Zones/Bins (11)
| Endpoint | Covered | Test Type | Test Files | Evidence |
|---|---|---|---|---|
| `POST /api/v1/warehouses` | yes | true no-mock HTTP | `API_tests/warehouse_crud_test.go` | `TestAdminWarehouseCRUD` |
| `GET /api/v1/warehouses` | yes | true no-mock HTTP | `API_tests/warehouse_test.go` | `TestUserCannotListWarehouses` |
| `GET /api/v1/warehouses/:id` | yes | true no-mock HTTP | `API_tests/warehouse_crud_test.go` | `TestAdminWarehouseCRUD` |
| `PUT /api/v1/warehouses/:id` | yes | true no-mock HTTP | `API_tests/warehouse_crud_test.go` | `TestAdminWarehouseCRUD` |
| `DELETE /api/v1/warehouses/:id` | yes | true no-mock HTTP | `API_tests/warehouse_crud_test.go` | `TestAdminWarehouseCRUD` |
| `POST /api/v1/warehouses/:id/zones` | yes | true no-mock HTTP | `API_tests/warehouse_crud_test.go` | `TestAdminWarehouseCRUD` |
| `GET /api/v1/warehouses/:id/zones` | yes | true no-mock HTTP | `API_tests/warehouse_crud_test.go` | `TestAdminWarehouseCRUD` |
| `POST /api/v1/zones/:id/bins` | yes | true no-mock HTTP | `API_tests/warehouse_crud_test.go` | `TestAdminWarehouseCRUD` |
| `GET /api/v1/zones/:id/bins` | yes | true no-mock HTTP | `API_tests/warehouse_crud_test.go` | `TestAdminWarehouseCRUD` |
| `PUT /api/v1/bins/:id` | yes | true no-mock HTTP | `API_tests/warehouse_crud_test.go` | `TestAdminWarehouseCRUD` |
| `DELETE /api/v1/bins/:id` | yes | true no-mock HTTP | `API_tests/warehouse_crud_test.go` | `TestAdminWarehouseCRUD` |

#### Categories/Items/Units (11)
| Endpoint | Covered | Test Type | Test Files | Evidence |
|---|---|---|---|---|
| `POST /api/v1/categories` | yes | true no-mock HTTP | `API_tests/item_crud_test.go` | `TestCategoryAndItemCRUD` |
| `GET /api/v1/categories` | yes | true no-mock HTTP | `API_tests/item_crud_test.go` | `TestCategoryAndItemCRUD` |
| `POST /api/v1/items` | yes | true no-mock HTTP | `API_tests/item_crud_test.go` | `TestCategoryAndItemCRUD` |
| `GET /api/v1/items` | yes | true no-mock HTTP | `API_tests/item_crud_test.go` | `TestCategoryAndItemCRUD` |
| `GET /api/v1/items/:id` | yes | true no-mock HTTP | `API_tests/item_crud_test.go` | `TestCategoryAndItemCRUD` |
| `PUT /api/v1/items/:id` | yes | true no-mock HTTP | `API_tests/item_crud_test.go` | `TestCategoryAndItemCRUD` |
| `DELETE /api/v1/items/:id` | yes | true no-mock HTTP | `API_tests/item_crud_test.go` | `TestCategoryAndItemCRUD` |
| `POST /api/v1/units` | yes | true no-mock HTTP | `API_tests/item_crud_test.go` | `TestUnitAndConversionCRUD` |
| `GET /api/v1/units` | yes | true no-mock HTTP | `API_tests/item_crud_test.go` | `TestUnitAndConversionCRUD` |
| `POST /api/v1/units/conversions` | yes | true no-mock HTTP | `API_tests/item_crud_test.go` | `TestUnitAndConversionCRUD` |
| `GET /api/v1/units/conversions` | yes | true no-mock HTTP | `API_tests/item_crud_test.go` | `TestUnitAndConversionCRUD` |

#### Suppliers/Carriers (8)
| Endpoint | Covered | Test Type | Test Files | Evidence |
|---|---|---|---|---|
| `POST /api/v1/suppliers` | yes | true no-mock HTTP | `API_tests/supplier_crud_test.go` | `TestSupplierCRUD` |
| `GET /api/v1/suppliers` | yes | true no-mock HTTP | `API_tests/supplier_crud_test.go` | `TestSupplierCRUD` |
| `GET /api/v1/suppliers/:id` | yes | true no-mock HTTP | `API_tests/supplier_crud_test.go` | `TestSupplierCRUD` |
| `PUT /api/v1/suppliers/:id` | yes | true no-mock HTTP | `API_tests/supplier_crud_test.go` | `TestSupplierCRUD` |
| `POST /api/v1/carriers` | yes | true no-mock HTTP | `API_tests/supplier_crud_test.go` | `TestCarrierCRUD` |
| `GET /api/v1/carriers` | yes | true no-mock HTTP | `API_tests/supplier_crud_test.go` | `TestCarrierCRUD` |
| `GET /api/v1/carriers/:id` | yes | true no-mock HTTP | `API_tests/supplier_crud_test.go` | `TestCarrierCRUD` |
| `PUT /api/v1/carriers/:id` | yes | true no-mock HTTP | `API_tests/supplier_crud_test.go` | `TestCarrierCRUD` |

#### Stations/Devices (9)
| Endpoint | Covered | Test Type | Test Files | Evidence |
|---|---|---|---|---|
| `POST /api/v1/stations` | yes | true no-mock HTTP | `API_tests/station_crud_test.go` | `TestAdminStationDeviceCRUD` |
| `GET /api/v1/stations` | yes | true no-mock HTTP | `API_tests/station_auth_test.go` | `TestUserCannotAccessStations` |
| `GET /api/v1/stations/:id` | yes | true no-mock HTTP | `API_tests/station_crud_test.go` | `TestAdminStationDeviceCRUD` |
| `PUT /api/v1/stations/:id` | yes | true no-mock HTTP | `API_tests/station_crud_test.go` | `TestAdminStationDeviceCRUD` |
| `DELETE /api/v1/stations/:id` | yes | true no-mock HTTP | `API_tests/station_crud_test.go` | `TestAdminStationDeviceCRUD` |
| `POST /api/v1/stations/:id/devices` | yes | true no-mock HTTP | `API_tests/station_crud_test.go` | `TestAdminStationDeviceCRUD` |
| `GET /api/v1/stations/:id/devices` | yes | true no-mock HTTP | `API_tests/station_crud_test.go` | `TestAdminStationDeviceCRUD` |
| `PUT /api/v1/devices/:id` | yes | true no-mock HTTP | `API_tests/station_crud_test.go` | `TestAdminStationDeviceCRUD` |
| `DELETE /api/v1/devices/:id` | yes | true no-mock HTTP | `API_tests/station_crud_test.go` | `TestAdminStationDeviceCRUD` |

#### Pricing (12)
| Endpoint | Covered | Test Type | Test Files | Evidence |
|---|---|---|---|---|
| `POST /api/v1/pricing/templates` | yes | true no-mock HTTP | `API_tests/pricing_crud_test.go` | `TestPricingLifecycle` |
| `GET /api/v1/pricing/templates` | yes | true no-mock HTTP | `API_tests/pricing_test.go` | `TestUserCannotAccessPricingTemplates` |
| `GET /api/v1/pricing/templates/:id` | yes | true no-mock HTTP | `API_tests/cross_org_access_test.go` | `TestMerchantCannotAccessOtherOrgPricingTemplate` |
| `POST /api/v1/pricing/templates/:id/versions` | yes | true no-mock HTTP | `API_tests/pricing_crud_test.go` | `TestPricingLifecycle` |
| `GET /api/v1/pricing/templates/:id/versions` | yes | true no-mock HTTP | `API_tests/pricing_crud_test.go` | `TestPricingLifecycle` |
| `GET /api/v1/pricing/versions/:id` | yes | true no-mock HTTP | `API_tests/order_pricing_test.go` | `TestFutureDatedActivation_PreservesCurrentPricing` |
| `POST /api/v1/pricing/versions/:id/activate` | yes | true no-mock HTTP | `API_tests/pricing_crud_test.go` | `TestPricingLifecycle` |
| `POST /api/v1/pricing/versions/:id/deactivate` | yes | true no-mock HTTP | `API_tests/pricing_crud_test.go` | `TestPricingLifecycle` |
| `POST /api/v1/pricing/versions/:id/rollback` | yes | true no-mock HTTP | `API_tests/pricing_crud_test.go` | `TestPricingLifecycle` |
| `POST /api/v1/pricing/versions/:id/tou-rules` | yes | true no-mock HTTP | `API_tests/pricing_crud_test.go` | `TestPricingLifecycle` |
| `GET /api/v1/pricing/versions/:id/tou-rules` | yes | true no-mock HTTP | `API_tests/pricing_crud_test.go` | `TestPricingLifecycle` |
| `DELETE /api/v1/pricing/tou-rules/:id` | yes | true no-mock HTTP | `API_tests/pricing_tou_delete_test.go` | `TestDeleteTOURule` |

#### Orders (4)
| Endpoint | Covered | Test Type | Test Files | Evidence |
|---|---|---|---|---|
| `POST /api/v1/orders` | yes | true no-mock HTTP | `API_tests/order_auth_test.go` | `TestUserCanCreateOrder_NoActivePricing` |
| `GET /api/v1/orders` | yes | true no-mock HTTP | `API_tests/order_auth_test.go` | `TestUserCanListOwnOrders` |
| `GET /api/v1/orders/:id` | yes | true no-mock HTTP | `API_tests/order_auth_test.go` | `TestOrderGetRequiresOwnership` |
| `POST /api/v1/orders/:id/recalculate` | yes | true no-mock HTTP | `API_tests/order_pricing_test.go` | `TestRecalculate_AfterDeactivation` |

#### Content (12)
| Endpoint | Covered | Test Type | Test Files | Evidence |
|---|---|---|---|---|
| `POST /api/v1/content/carousel` | yes | true no-mock HTTP | `API_tests/content_crud_test.go` | `TestCarouselCRUD` |
| `GET /api/v1/content/carousel` | yes | true no-mock HTTP | `API_tests/content_crud_test.go` | `TestCarouselCRUD` |
| `PUT /api/v1/content/carousel/:id` | yes | true no-mock HTTP | `API_tests/content_crud_test.go` | `TestCarouselCRUD` |
| `DELETE /api/v1/content/carousel/:id` | yes | true no-mock HTTP | `API_tests/content_crud_test.go` | `TestCarouselCRUD` |
| `POST /api/v1/content/campaigns` | yes | true no-mock HTTP | `API_tests/content_crud_test.go` | `TestCampaignCRUD` |
| `GET /api/v1/content/campaigns` | yes | true no-mock HTTP | `API_tests/content_crud_test.go` | `TestCampaignCRUD` |
| `PUT /api/v1/content/campaigns/:id` | yes | true no-mock HTTP | `API_tests/content_crud_test.go` | `TestCampaignCRUD` |
| `DELETE /api/v1/content/campaigns/:id` | yes | true no-mock HTTP | `API_tests/content_crud_test.go` | `TestCampaignCRUD` |
| `POST /api/v1/content/rankings` | yes | true no-mock HTTP | `API_tests/content_crud_test.go` | `TestRankingCRUD` |
| `GET /api/v1/content/rankings` | yes | true no-mock HTTP | `API_tests/content_crud_test.go` | `TestRankingCRUD` |
| `PUT /api/v1/content/rankings/:id` | yes | true no-mock HTTP | `API_tests/content_crud_test.go` | `TestRankingCRUD` |
| `DELETE /api/v1/content/rankings/:id` | yes | true no-mock HTTP | `API_tests/content_crud_test.go` | `TestRankingCRUD` |

#### Notifications (9)
| Endpoint | Covered | Test Type | Test Files | Evidence |
|---|---|---|---|---|
| `GET /api/v1/notifications/inbox` | yes | true no-mock HTTP | `API_tests/notification_auth_test.go` | `TestUserCanListOwnInbox` |
| `POST /api/v1/notifications/inbox/:id/read` | yes | true no-mock HTTP | `API_tests/notification_policy_test.go` | `TestNotificationMarkReadAndDismiss` |
| `POST /api/v1/notifications/inbox/:id/dismiss` | yes | true no-mock HTTP | `API_tests/notification_policy_test.go` | `TestNotificationMarkReadAndDismiss` |
| `GET /api/v1/notifications/subscriptions` | yes | true no-mock HTTP | `API_tests/notification_auth_test.go` | `TestUserCanListSubscriptions` |
| `PUT /api/v1/notifications/subscriptions/:id` | yes | true no-mock HTTP | `API_tests/notification_crud_test.go` | `TestNotificationFlow` |
| `POST /api/v1/notifications/templates` | yes | true no-mock HTTP | `API_tests/notification_crud_test.go` | `TestNotificationFlow` |
| `GET /api/v1/notifications/templates` | yes | true no-mock HTTP | `API_tests/notification_crud_test.go` | `TestNotificationFlow` |
| `POST /api/v1/notifications/send` | yes | true no-mock HTTP | `API_tests/notification_policy_test.go` | `TestNotificationDelivery_WorkerVerified` |
| `GET /api/v1/notifications/stats` | yes | true no-mock HTTP | `API_tests/notification_crud_test.go` | `TestNotificationFlow` |

#### Admin (5)
| Endpoint | Covered | Test Type | Test Files | Evidence |
|---|---|---|---|---|
| `GET /api/v1/admin/audit-logs` | yes | true no-mock HTTP | `API_tests/admin_crud_test.go` | `TestAdminCanViewAuditLogs` |
| `GET /api/v1/admin/config` | yes | true no-mock HTTP | `API_tests/admin_crud_test.go` | `TestAdminCanViewConfig` |
| `PUT /api/v1/admin/config/:key` | yes | true no-mock HTTP | `API_tests/admin_crud_test.go` | `TestAdminCanUpdateConfig` |
| `GET /api/v1/admin/metrics` | yes | true no-mock HTTP | `API_tests/admin_crud_test.go` | `TestAdminCanViewMetrics` |
| `GET /api/v1/admin/metrics/export` | yes | true no-mock HTTP | `API_tests/admin_crud_test.go` | `TestAdminCanExportMetricsCSV` |

### Coverage Summary
- Total endpoints: **102**
- Endpoints with HTTP tests: **102**
- Endpoints with TRUE no-mock tests: **102**
- HTTP coverage: **100%**
- True API coverage: **100%**

### Unit Test Summary

#### Backend Unit Tests
Files:
- `repo/unit_tests/apperror_test.go`
- `repo/unit_tests/auth_helpers_test.go`
- `repo/unit_tests/config_test.go`
- `repo/unit_tests/masking_test.go`
- `repo/unit_tests/notification_worker_test.go`
- `repo/unit_tests/password_test.go`
- `repo/unit_tests/pricing_logic_test.go`

Modules covered:
- Controllers/handlers: partial (`auth_helpers_test.go`)
- Services: pricing and notification worker behavior
- Repositories: no direct repository unit suite
- Auth/guards/middleware/error handling: covered (`auth_helpers_test.go`, `apperror_test.go`)

Important backend modules NOT directly tested:
- `repo/internal/repo/*`
- additional isolated service internals beyond API-path assertions

#### Frontend Unit Tests (STRICT)
- Frontend test files: **NONE**
- Frameworks/tools detected: **NONE**
- Components/modules covered: **NONE**
- Important frontend components/modules not tested: N/A (frontend not detected)

**Frontend unit tests: MISSING**

Strict failure rule:
- Project inferred as backend, so no frontend CRITICAL GAP trigger.

### API Observability Check
- Strong visibility of method/path/body in request calls.
- Strong response assertions in many CRUD/policy tests.
- Some tests remain mostly status-code oriented.

### Test Quality & Sufficiency
- Success paths: strong across domains.
- Failure/auth/validation paths: strong.
- Edge cases: lockout, device mismatch, TOU overlap/delete, quiet hours, rate limiting.
- Integration boundaries: real HTTP exercised.
- Over-mocking: not observed in API tests.

`repo/run_tests.sh` static check:
- Docker-based execution (`docker compose ... go test ...`) -> OK
- No local npm/pip/apt dependency in test flow -> OK

### Tests Check
- Coverage is high and materially confidence-building.
- Strict endpoint completeness requirement is satisfied (all shipped endpoints mapped to real HTTP test evidence).

### Test Coverage Score (0-100)
**93/100**

### Score Rationale
- Full true no-mock HTTP endpoint coverage with strong behavioral breadth.
- Score reduced for weaker direct repository-layer unit isolation and some assertion-light tests.

### Key Gaps
- Repository layer still lacks direct unit tests.
- A subset of tests remain status-focused instead of deep semantic payload assertions.

### Confidence & Assumptions
- Confidence: high.
- Static-only audit.
- Coverage conclusion is based on route wiring + visible API test request calls.

---

## 2) README Audit

### README Location
- Required: `repo/README.md`
- Found: `repo/README.md`

### Hard Gate Checks

#### Formatting
- PASS

#### Startup Instructions
- PASS (`docker-compose up` is present)

#### Access Method
- PASS (URL and port documented)

#### Verification Method
- PASS (curl verification flow included)

#### Environment Rules (STRICT)
- PASS
- No npm/pip/apt/runtime install instructions.
- No manual DB bootstrap command required in README flow.

#### Demo Credentials (Conditional)
- PASS
- Explicit credential matrix for all roles is present.

### Engineering Quality
- Tech stack clarity: strong
- Architecture explanation: strong
- Testing instructions: strong
- Security/roles/workflows: strong
- Presentation quality: strong

### High Priority Issues
- None.

### Medium Priority Issues
- README claims `101/101` endpoint coverage, while static route inventory shows `102` endpoints in `cmd/server/main.go`.

### Low Priority Issues
- None material.

### Hard Gate Failures
- None.

### README Verdict
**PASS**

---

## Final Combined Verdict
- **Test Coverage Audit Verdict:** PASS (strict endpoint completeness achieved; minor depth gaps remain)
- **README Audit Verdict:** PASS
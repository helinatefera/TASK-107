# ChargeOps API Specification (Implementation-Derived)

## Runtime Reality (Important)

- This API is a live HTTP server implemented in Go (Echo) with PostgreSQL persistence.
- This document is generated from implementation in repo and is the contract of current behavior.
- Validation, authorization, masking, and error payloads are enforced in middleware/handlers/services.

## Source of Truth Used

- Route wiring: `repo/cmd/server/main.go`
- Request DTOs: `repo/internal/dto/*.go`
- Response models: `repo/internal/model/*.go`
- Error contract: `repo/internal/apperror/errors.go`
- Auth and RBAC middleware: `repo/internal/middleware/*.go`
- Business rules: `repo/internal/service/*.go`
- Background behavior: `repo/internal/worker/*.go`
- Role and permission seeds: `repo/migrations/001_users_and_auth.up.sql`, `repo/migrations/011_user_read_permission_for_roles.up.sql`

## Contract Conventions

- Base URL: `http://localhost:8080`
- API prefix: `/api/v1`
- Request/response content type: `application/json` unless otherwise noted.
- Error envelope is always:

```json
{
	"code": 400,
	"msg": "validation failed: ..."
}
```

- IDs are UUID strings.
- Timestamps are RFC3339 strings in responses.
- Decimal money/rate fields are serialized by `shopspring/decimal` as JSON strings.

## Global Headers

- `Authorization: Bearer <token>` is required for all authenticated endpoints.
- `X-Device-Id: <device-id>` is required for all authenticated endpoints.
	- Missing or mismatched `X-Device-Id` returns `401` (`device mismatch`).
- `X-Request-Id` is optional on request and always set in response.

## Authentication and Session Model

- Public (no token needed):
	- `POST /api/v1/auth/register`
	- `POST /api/v1/auth/login`
	- `POST /api/v1/auth/recover`
	- `POST /api/v1/auth/recover/reset`
	- `GET /health`
- Guest-only unauthenticated read access:
	- `GET /api/v1/content/carousel`
	- `GET /api/v1/content/campaigns`
	- `GET /api/v1/content/rankings`
- Session controls:
	- Idle timeout: 30 minutes
	- Absolute timeout: 7 days
	- Refresh extends idle timeout only
- Password policy:
	- At least 12 chars
	- At least 3 of 4 classes: lowercase, uppercase, digits, special
- Login lockout policy:
	- Window: 10 minutes
	- Threshold: 10 failed attempts
	- Lock duration: 15 minutes
- Recovery token:
	- One-time use, expires in 1 hour
	- Stored as hash

## Roles and Permissions

Roles:

- `guest`
- `user`
- `merchant`
- `administrator`

Permissions seeded in DB:

- `user.read`, `user.manage`
- `org.read`, `org.manage`
- `warehouse.read`, `warehouse.manage`
- `item.read`, `item.manage`
- `supplier.read`, `supplier.manage`
- `station.read`, `station.manage`
- `pricing.read`, `pricing.manage`
- `order.read`, `order.create`
- `content.read`, `content.manage`
- `notification.read`, `notification.manage`
- `admin.config`, `admin.audit`, `admin.metrics`

Notes:

- Administrators are granted all permissions by seed migration.
- `user.read` is explicitly granted to `user` and `merchant` by migration 011.
- Route access requires both role checks and permission checks where configured.

## Shared Query Patterns

Many list endpoints support:

- `limit` (default `50` when missing or invalid/<=0)
- `offset` (default `0` when missing or invalid/<0)

## Error Codes Used by Implementation

- `400`: bad request, validation failure, invalid IDs, invalid time ranges
- `401`: unauthorized, session expired, device mismatch
- `403`: forbidden
- `404`: resource not found
- `409`: conflict / duplicate
- `429`: account locked
- `500`: internal server error

---

## Endpoint Catalog

### Health

#### `GET /health`

- Auth: none
- Response `200`:

```json
{"status":"ok"}
```

---

### Auth

#### `POST /api/v1/auth/register`

- Auth: none
- Body:

```json
{
	"email": "user@example.com",
	"password": "StrongPass123!",
	"display_name": "User"
}
```

- Response `201`: `User`

#### `POST /api/v1/auth/login`

- Auth: none
- Body:

```json
{
	"email": "user@example.com",
	"password": "StrongPass123!",
	"device_id": "my-device"
}
```

- Response `200`:

```json
{
	"token": "<bearer-token>",
	"expires_at": "2026-04-08T10:00:00Z"
}
```

#### `POST /api/v1/auth/logout`

- Auth: bearer + `X-Device-Id`
- Response `204`

#### `POST /api/v1/auth/refresh`

- Auth: bearer + `X-Device-Id`
- Body:

```json
{
	"device_id": "my-device"
}
```

- Response `200`:

```json
{"status":"refreshed"}
```

#### `POST /api/v1/auth/recover`

- Auth: none
- Body:

```json
{"email":"user@example.com"}
```

- Response `200`:

```json
{"token":"<recovery-token>"}
```

#### `POST /api/v1/auth/recover/reset`

- Auth: none
- Body:

```json
{
	"token": "<recovery-token>",
	"new_password": "AnotherStrongPass123!"
}
```

- Response `200`:

```json
{"status":"password_reset"}
```

---

### Users and Permissions

#### `GET /api/v1/users`

- Role/permission: administrator + `user.read`
- Query: `limit`, `offset`
- Response `200`: `User[]`

#### `GET /api/v1/users/me`

- Role/permission: administrator|merchant|user + `user.read`
- Response `200`: `User`

#### `GET /api/v1/users/:id`

- Role/permission: administrator|merchant|user + `user.read`
- Additional rule: self-or-admin only
- Response `200`: `User`

#### `PUT /api/v1/users/:id`

- Role/permission: administrator|merchant|user + `user.read`
- Additional rule: self-or-admin only
- Body:

```json
{
	"display_name": "New Name",
	"email": "new@example.com"
}
```

- Response `200`: `User`

#### `PUT /api/v1/users/:id/role`

- Role/permission: administrator + `user.manage`
- Body:

```json
{"role":"merchant"}
```

- Response `200`: `{"status":"role_updated"}`

#### `PUT /api/v1/users/:id/org`

- Role/permission: administrator + `user.manage`
- Body:

```json
{"org_id":"<uuid-or-null>"}
```

- Response `200`: `{"status":"org_updated"}`

#### `DELETE /api/v1/users/:id`

- Role/permission: administrator + `user.manage`
- Response `204`

#### `GET /api/v1/users/:id/permissions`

- Role/permission: administrator|merchant|user + `user.read`
- Additional rule: self-or-admin only
- Response `200`: `Permission[]`

#### `PUT /api/v1/users/:id/permissions`

- Role/permission: administrator + `user.manage`
- Body:

```json
{
	"permissions": [
		{"permission_id":"<uuid>","granted":true}
	]
}
```

- Response `200`: `{"status":"permissions_updated"}`

---

### Organizations

#### `POST /api/v1/orgs`

- Role/permission: administrator + `org.manage`
- Body:

```json
{
	"parent_id": null,
	"org_code": "ORG-001",
	"name": "Org Name",
	"tax_id": "123-45-6789",
	"address": "Address",
	"timezone": "UTC"
}
```

- Response `201`: `OrgResponse`

#### `GET /api/v1/orgs`

- Role/permission: administrator|merchant + `org.read`
- Query: `limit`, `offset`
- Scope:
	- administrator: all orgs
	- merchant: own org and descendants
- Response `200`: `OrgResponse[]`

#### `GET /api/v1/orgs/:id`

- Role/permission: administrator|merchant + `org.read`
- Scope: org ownership/descendant check for non-admin
- Response `200`: `OrgResponse`

#### `PUT /api/v1/orgs/:id`

- Role/permission: administrator + `org.manage`
- Body: `UpdateOrgRequest`
- Response `200`: `OrgResponse`

#### `DELETE /api/v1/orgs/:id`

- Role/permission: administrator + `org.manage`
- Response `204`

---

### Warehouses, Zones, Bins

#### `POST /api/v1/warehouses`

- Role/permission: administrator|merchant + `warehouse.manage`
- Body: `CreateWarehouseRequest`
- Merchant behavior: request `org_id` is overridden to caller org
- Response `201`: `Warehouse` (address masked for non-admin)

#### `GET /api/v1/warehouses`

- Role/permission: administrator|merchant + `warehouse.read`
- Query: `limit`, `offset`
- Response `200`: `Warehouse[]` (address masked for non-admin)

#### `GET /api/v1/warehouses/:id`

- Role/permission: administrator|merchant + `warehouse.read`
- Response `200`: `Warehouse` (masked for non-admin)

#### `PUT /api/v1/warehouses/:id`

- Role/permission: administrator|merchant + `warehouse.manage`
- Body: `UpdateWarehouseRequest`
- Response `200`: `Warehouse` (masked for non-admin)

#### `DELETE /api/v1/warehouses/:id`

- Role/permission: administrator + `warehouse.manage`
- Response `204`

#### `POST /api/v1/warehouses/:id/zones`

- Role/permission: administrator|merchant + `warehouse.manage`
- Body:

```json
{"name":"Zone A","zone_type":"cold"}
```

- Response `201`: `Zone`

#### `GET /api/v1/warehouses/:id/zones`

- Role/permission: administrator|merchant + `warehouse.read`
- Response `200`: `Zone[]`

#### `POST /api/v1/zones/:id/bins`

- Role/permission: administrator|merchant + `warehouse.manage`
- Body:

```json
{"bin_code":"B-100","capacity":100}
```

- Response `201`: `Bin`

#### `GET /api/v1/zones/:id/bins`

- Role/permission: administrator|merchant + `warehouse.read`
- Response `200`: `Bin[]`

#### `PUT /api/v1/bins/:id`

- Role/permission: administrator|merchant + `warehouse.manage`
- Body: `UpdateBinRequest`
- Response `200`: `Bin`

#### `DELETE /api/v1/bins/:id`

- Role/permission: administrator + `warehouse.manage`
- Response `204`

---

### Categories, Items, Units

#### `POST /api/v1/categories`

- Role/permission: administrator + `item.manage`
- Body: `CreateCategoryRequest`
- Response `201`: `Category`

#### `GET /api/v1/categories`

- Role/permission: administrator|merchant|user + `item.read`
- Response `200`: `Category[]`

#### `POST /api/v1/items`

- Role/permission: administrator|merchant + `item.manage`
- Body: `CreateItemRequest`
- Response `201`: `Item`

#### `GET /api/v1/items`

- Role/permission: administrator|merchant|user + `item.read`
- Query: `limit`, `offset`
- Response `200`: `Item[]`

#### `GET /api/v1/items/:id`

- Role/permission: administrator|merchant|user + `item.read`
- Response `200`: `Item`

#### `PUT /api/v1/items/:id`

- Role/permission: administrator|merchant + `item.manage`
- Body: `UpdateItemRequest`
- Response `200`: `Item`

#### `DELETE /api/v1/items/:id`

- Role/permission: administrator + `item.manage`
- Response `204`

#### `POST /api/v1/units`

- Role/permission: administrator + `item.manage`
- Body: `CreateUnitRequest`
- Response `201`: `UnitOfMeasure`

#### `GET /api/v1/units`

- Role/permission: administrator|merchant|user + `item.read`
- Response `200`: `UnitOfMeasure[]`

#### `POST /api/v1/units/conversions`

- Role/permission: administrator + `item.manage`
- Body: `CreateConversionRequest`
- Rule: `factor` must be positive
- Response `201`: `UnitConversion`

#### `GET /api/v1/units/conversions`

- Role/permission: administrator|merchant|user + `item.read`
- Response `200`: `UnitConversion[]`

---

### Suppliers and Carriers

#### `POST /api/v1/suppliers`

- Role/permission: administrator|merchant + `supplier.manage`
- Body: `CreateSupplierRequest`
- Merchant behavior: request `org_id` overridden to caller org
- Dedup rule: normalized `name` + `tax_id` (org-scoped)
- Response `201`: `Supplier` (sensitive fields masked for non-admin)

#### `GET /api/v1/suppliers`

- Role/permission: administrator|merchant + `supplier.read`
- Query: `limit`, `offset`
- Response `200`: `Supplier[]` (masked for non-admin)

#### `GET /api/v1/suppliers/:id`

- Role/permission: administrator|merchant + `supplier.read`
- Response `200`: `Supplier` (masked for non-admin)

#### `PUT /api/v1/suppliers/:id`

- Role/permission: administrator|merchant + `supplier.manage`
- Body: `UpdateSupplierRequest`
- Response `200`: `Supplier` (masked for non-admin)

#### `POST /api/v1/carriers`

- Role/permission: administrator|merchant + `supplier.manage`
- Body: `CreateCarrierRequest`
- Merchant behavior: request `org_id` overridden to caller org
- Dedup rule: normalized `name` + `tax_id` (org-scoped)
- Response `201`: `Carrier` (tax_id masked for non-admin)

#### `GET /api/v1/carriers`

- Role/permission: administrator|merchant + `supplier.read`
- Query: `limit`, `offset`
- Response `200`: `Carrier[]`

#### `GET /api/v1/carriers/:id`

- Role/permission: administrator|merchant + `supplier.read`
- Response `200`: `Carrier`

#### `PUT /api/v1/carriers/:id`

- Role/permission: administrator|merchant + `supplier.manage`
- Body: `UpdateCarrierRequest`
- Response `200`: `Carrier`

---

### Stations and Devices

#### `POST /api/v1/stations`

- Role/permission: administrator|merchant + `station.manage`
- Body: `CreateStationRequest`
- Merchant behavior: request `org_id` overridden to caller org
- Response `201`: `Station`

#### `GET /api/v1/stations`

- Role/permission: administrator|merchant + `station.read`
- Query: `limit`, `offset`
- Response `200`: `Station[]`

#### `GET /api/v1/stations/:id`

- Role/permission: administrator|merchant + `station.read`
- Response `200`: `Station`

#### `PUT /api/v1/stations/:id`

- Role/permission: administrator|merchant + `station.manage`
- Body: `UpdateStationRequest`
- Response `200`: `Station`

#### `DELETE /api/v1/stations/:id`

- Role/permission: administrator + `station.manage`
- Response `204`

#### `POST /api/v1/stations/:id/devices`

- Role/permission: administrator|merchant + `station.manage`
- Body: `CreateDeviceRequest`
- Response `201`: `Device`

#### `GET /api/v1/stations/:id/devices`

- Role/permission: administrator|merchant + `station.read`
- Response `200`: `Device[]`

#### `PUT /api/v1/devices/:id`

- Role/permission: administrator|merchant + `station.manage`
- Body: `UpdateDeviceRequest`
- Response `200`: `Device`

#### `DELETE /api/v1/devices/:id`

- Role/permission: administrator|merchant + `station.manage`
- Response `204`

---

### Pricing

#### `POST /api/v1/pricing/templates`

- Role/permission: administrator|merchant + `pricing.manage`
- Body: `CreatePriceTemplateRequest`
- Rule: exactly one of `station_id` or `device_id` must be set
- Response `201`: `PriceTemplate`

#### `GET /api/v1/pricing/templates`

- Role/permission: administrator|merchant + `pricing.read`
- Query: `limit`, `offset`
- Response `200`: `PriceTemplate[]`

#### `GET /api/v1/pricing/templates/:id`

- Role/permission: administrator|merchant + `pricing.read`
- Response `200`: `PriceTemplate`

#### `POST /api/v1/pricing/templates/:id/versions`

- Role/permission: administrator|merchant + `pricing.manage`
- Body: `CreateVersionRequest`
- Rules:
	- rates/fees must be non-negative
	- if `apply_tax=true`, tax rate is resolved from `app_config.tax_rate`
- Response `201`: `PriceTemplateVersion`

#### `GET /api/v1/pricing/templates/:id/versions`

- Role/permission: administrator|merchant + `pricing.read`
- Response `200`: `PriceTemplateVersion[]`

#### `GET /api/v1/pricing/versions/:id`

- Role/permission: administrator|merchant + `pricing.read`
- Response `200`: `PriceTemplateVersion`

#### `POST /api/v1/pricing/versions/:id/activate`

- Role/permission: administrator|merchant + `pricing.manage`
- Body (optional):

```json
{"effective_at":"2026-04-08T12:00:00Z"}
```

- Rules:
	- if omitted, `effective_at=now`
	- explicit `effective_at` cannot be in past
	- future activation keeps current active version to avoid pricing gap
- Response `200`: `PriceTemplateVersion`

#### `POST /api/v1/pricing/versions/:id/deactivate`

- Role/permission: administrator|merchant + `pricing.manage`
- Response `200`:

```json
{"status":"deactivated"}
```

#### `POST /api/v1/pricing/versions/:id/rollback`

- Role/permission: administrator|merchant + `pricing.manage`
- Behavior: clones selected version and its TOU rules into a new draft version
- Response `201`: `PriceTemplateVersion`

#### `POST /api/v1/pricing/versions/:id/tou-rules`

- Role/permission: administrator|merchant + `pricing.manage`
- Body: `CreateTOURuleRequest`
- Rules:
	- `day_type` in `weekday|weekend|holiday`
	- `start_time < end_time`
	- non-negative rates
	- no overlaps with existing rules in same version/day_type
- Response `201`: `TOURule`

#### `GET /api/v1/pricing/versions/:id/tou-rules`

- Role/permission: administrator|merchant + `pricing.read`
- Response `200`: `TOURule[]`

#### `DELETE /api/v1/pricing/tou-rules/:id`

- Role/permission: administrator|merchant + `pricing.manage`
- Response `204`

---

### Orders

#### `POST /api/v1/orders`

- Role/permission: administrator|merchant|user + `order.create`
- Body: `CreateOrderRequest`
- Additional validations:
	- `end_time` must be after `start_time`
	- `energy_kwh` must be positive
- Pricing behavior:
	- resolves active version by device first, then station fallback
	- applies TOU rule based on station timezone/day type
	- duration is ceil-rounded minutes
	- snapshot stores all pricing components immutably
- Response `201`: `OrderSnapshot`

#### `GET /api/v1/orders`

- Role/permission: administrator|merchant|user + `order.read`
- Query: `limit`, `offset`
- Scope:
	- admin: all
	- merchant/user: own orders only
- Response `200`: `OrderSnapshot[]`

#### `GET /api/v1/orders/:id`

- Role/permission: administrator|merchant|user + `order.read`
- Scope: non-admin can read only own order
- Response `200`: `OrderSnapshot`

#### `POST /api/v1/orders/:id/recalculate`

- Role/permission: administrator + `order.create`
- Behavior: recalculates from original stored version and order window, does not persist
- Response `200`: recalculated `OrderSnapshot` payload

---

### Content (Carousel, Campaigns, Rankings)

Public guest reads are supported for all list endpoints below.

#### Carousel

- `POST /api/v1/content/carousel`
	- Role/permission: administrator|merchant + `content.manage`
	- Body: `CreateCarouselRequest`
	- Response `201`: `CarouselSlot`
- `GET /api/v1/content/carousel`
	- Role/permission: administrator|merchant|user|guest + `content.read`
	- Response `200`: `CarouselSlot[]`
- `PUT /api/v1/content/carousel/:id`
	- Role/permission: administrator|merchant + `content.manage`
	- Body: `UpdateCarouselRequest`
	- Response `200`: `CarouselSlot`
- `DELETE /api/v1/content/carousel/:id`
	- Role/permission: administrator + `content.manage`
	- Response `204`

#### Campaigns

- `POST /api/v1/content/campaigns`
	- Role/permission: administrator|merchant + `content.manage`
	- Body: `CreateCampaignRequest`
	- Response `201`: `CampaignPlacement`
- `GET /api/v1/content/campaigns`
	- Role/permission: administrator|merchant|user|guest + `content.read`
	- Response `200`: `CampaignPlacement[]`
- `PUT /api/v1/content/campaigns/:id`
	- Role/permission: administrator|merchant + `content.manage`
	- Body: `UpdateCampaignRequest`
	- Response `200`: `CampaignPlacement`
- `DELETE /api/v1/content/campaigns/:id`
	- Role/permission: administrator + `content.manage`
	- Response `204`

#### Rankings

- `POST /api/v1/content/rankings`
	- Role/permission: administrator|merchant + `content.manage`
	- Body: `CreateRankingRequest`
	- Response `201`: `HotRanking`
- `GET /api/v1/content/rankings`
	- Role/permission: administrator|merchant|user|guest + `content.read`
	- Response `200`: `HotRanking[]`
- `PUT /api/v1/content/rankings/:id`
	- Role/permission: administrator|merchant + `content.manage`
	- Body: `UpdateRankingRequest`
	- Response `200`: `HotRanking`
- `DELETE /api/v1/content/rankings/:id`
	- Role/permission: administrator + `content.manage`
	- Response `204`

---

### Notifications

#### Inbox

#### `GET /api/v1/notifications/inbox`

- Role/permission: administrator|merchant|user + `notification.read`
- Query: `limit`, `offset`
- Response `200`: `Message[]`

#### `POST /api/v1/notifications/inbox/:id/read`

- Role/permission: administrator|merchant|user + `notification.read`
- Rule: caller must own message
- Response `200`: `{"status":"read"}`

#### `POST /api/v1/notifications/inbox/:id/dismiss`

- Role/permission: administrator|merchant|user + `notification.read`
- Rule: caller must own message
- Response `200`: `{"status":"dismissed"}`

#### Subscriptions

#### `GET /api/v1/notifications/subscriptions`

- Role/permission: administrator|merchant|user + `notification.read`
- Response `200`: `NotificationSubscription[]`

#### `PUT /api/v1/notifications/subscriptions/:id`

- Role/permission: administrator|merchant|user + `notification.read`
- Body:

```json
{"opted_out":true}
```

- Response `200`: `{"status":"updated"}`

#### Templates and Sending

#### `POST /api/v1/notifications/templates`

- Role/permission: administrator + `notification.manage`
- Body: `CreateNotificationTemplateRequest`
- Response `201`: `NotificationTemplate`

#### `GET /api/v1/notifications/templates`

- Role/permission: administrator + `notification.manage`
- Response `200`: `NotificationTemplate[]`

#### `POST /api/v1/notifications/send`

- Role/permission: administrator + `notification.manage`
- Body: `SendNotificationRequest`
- Response `201`: `{"status":"queued"}`

#### `GET /api/v1/notifications/stats`

- Role/permission: administrator + `notification.manage`
- Query: optional `template_id`
- Response `200`: `DeliveryStats[]`

---

### Admin

#### `GET /api/v1/admin/audit-logs`

- Role/permission: administrator + `admin.audit`
- Query:
	- `entity_type` optional
	- `entity_id` optional
	- `limit`, `offset`
- Response `200`: `AuditLog[]`

#### `GET /api/v1/admin/config`

- Role/permission: administrator + `admin.config`
- Response `200`: `AppConfig[]`

#### `PUT /api/v1/admin/config/:key`

- Role/permission: administrator + `admin.config`
- Body:

```json
{"value":"0.08"}
```

- Validation:
	- For `tax_rate`, value must parse float and be within `[0,1]`
- Response `200`: `AppConfig`

#### `GET /api/v1/admin/metrics`

- Role/permission: administrator + `admin.metrics`
- Query: `limit`, `offset`
- Response `200`: `RequestMetric[]`

#### `GET /api/v1/admin/metrics/export`

- Role/permission: administrator + `admin.metrics`
- Query:
	- `format`: `csv` (default) or `json`
	- `since`: RFC3339 (default: now - 24h)
	- `until`: RFC3339 optional
	- `path`: optional exact path filter
	- `status_code`: optional int filter
- Response `200`:
	- `text/csv` attachment (`metrics.csv`) or
	- `application/json` attachment (`metrics.json`)

---

## Background Worker Behavior (Operational Contract)

### Notification Worker

- Poll interval: every 5 seconds
- Batch claim size: 50 pending jobs
- Job timeout: 15 seconds
- Suppression and scheduling:
	- suppress if user opted out
	- defer during quiet hours (`21:00-06:59` local org timezone)
	- suppress when delivered count reaches 2 per template per local day
- Delivery channel: inserts into `messages` inbox table
- Template rendering: `{{key}}` placeholder replacement from JSON params

### Cleanup Worker

- Poll interval: every 10 minutes
- Deletes:
	- sessions past absolute expiry
	- used/expired recovery tokens
	- request metrics older than 30 days

---

## Core Schemas

### Error

```json
{
	"code": 400,
	"msg": "bad request"
}
```

### User

```json
{
	"id": "uuid",
	"email": "user@example.com",
	"display_name": "Name",
	"role": "user",
	"org_id": "uuid-or-null",
	"created_at": "2026-04-08T10:00:00Z",
	"updated_at": "2026-04-08T10:00:00Z"
}
```

### OrgResponse

```json
{
	"id": "uuid",
	"parent_id": "uuid-or-null",
	"org_code": "ORG-001",
	"name": "Org Name",
	"tax_id": "masked-or-full",
	"address": "masked-or-full",
	"timezone": "UTC",
	"created_at": "2026-04-08T10:00:00Z",
	"updated_at": "2026-04-08T10:00:00Z"
}
```

### Warehouse

```json
{
	"id": "uuid",
	"org_id": "uuid",
	"name": "Warehouse A",
	"address": "masked-or-full",
	"timezone": "UTC",
	"created_at": "2026-04-08T10:00:00Z",
	"updated_at": "2026-04-08T10:00:00Z"
}
```

### Item

```json
{
	"id": "uuid",
	"sku": "SKU-001",
	"item_name": "Cable",
	"category_id": "uuid-or-null",
	"base_unit_id": "uuid",
	"description": "optional",
	"created_at": "2026-04-08T10:00:00Z",
	"updated_at": "2026-04-08T10:00:00Z"
}
```

### Supplier / Carrier

```json
{
	"id": "uuid",
	"org_id": "uuid-or-null",
	"name": "Supplier Name",
	"tax_id": "masked-or-full",
	"contact_email": "optional@example.com",
	"address": "masked-or-full",
	"created_at": "2026-04-08T10:00:00Z",
	"updated_at": "2026-04-08T10:00:00Z"
}
```

### Station / Device

```json
{
	"id": "uuid",
	"org_id": "uuid",
	"name": "Station A",
	"location": "optional",
	"timezone": "UTC",
	"created_at": "2026-04-08T10:00:00Z",
	"updated_at": "2026-04-08T10:00:00Z"
}
```

```json
{
	"id": "uuid",
	"station_id": "uuid",
	"device_code": "DEV-001",
	"device_type": "optional",
	"status": "active",
	"created_at": "2026-04-08T10:00:00Z",
	"updated_at": "2026-04-08T10:00:00Z"
}
```

### PriceTemplate

```json
{
	"id": "uuid",
	"org_id": "uuid",
	"name": "Template A",
	"station_id": "uuid-or-null",
	"device_id": "uuid-or-null",
	"created_at": "2026-04-08T10:00:00Z"
}
```

### PriceTemplateVersion

```json
{
	"id": "uuid",
	"template_id": "uuid",
	"version_number": 1,
	"energy_rate": "0.20",
	"duration_rate": "0.05",
	"service_fee": "2.00",
	"apply_tax": true,
	"tax_rate": "0.08",
	"status": "active",
	"effective_at": "2026-04-08T10:00:00Z",
	"cloned_from_version_id": "uuid-or-null",
	"created_at": "2026-04-08T10:00:00Z"
}
```

### TOURule

```json
{
	"id": "uuid",
	"version_id": "uuid",
	"day_type": "weekday",
	"start_time": "09:00",
	"end_time": "18:00",
	"energy_rate": "0.25",
	"duration_rate": "0.06"
}
```

### OrderSnapshot

```json
{
	"id": "uuid",
	"order_id": "ORD-001",
	"user_id": "uuid",
	"device_id": "uuid",
	"station_id": "uuid",
	"version_id": "uuid",
	"energy_rate": "0.25",
	"duration_rate": "0.06",
	"service_fee": "2.00",
	"tax_rate": "0.08",
	"tou_applied": {},
	"energy_kwh": "12.5",
	"duration_min": 46,
	"subtotal": "8.00",
	"tax_amount": "0.64",
	"total": "8.64",
	"order_start": "2026-04-08T10:00:00Z",
	"order_end": "2026-04-08T10:45:00Z",
	"created_at": "2026-04-08T10:46:00Z"
}
```

### Content Models

- `CarouselSlot`
- `CampaignPlacement`
- `HotRanking`

All include:

- `id`, `org_id`, `priority`, `target_role`, `start_time`, `end_time`
- activation flags and timestamps (`active`, `created_at`, `updated_at` where applicable)

### Notification Models

- `NotificationTemplate`
- `NotificationSubscription`
- `Message`
- `DeliveryStats`

### Admin Models

- `AuditLog`
- `AppConfig`
- `RequestMetric`

## Auditing

- Middleware records successful mutating requests (`POST`, `PUT`, `DELETE`) for audited endpoints.
- Audit payload captures:
	- actor user id
	- action (`METHOD route`)
	- entity type/id
	- old and new values when handlers provide them
	- IP address and request id

## Masking Rules

- Non-admin users receive masked values:
	- tax IDs: only last 4 characters visible
	- addresses: fully masked (`****`)
- Applies to orgs, suppliers, carriers, warehouses (as implemented by handlers/services).

## Notes on Contract Completeness

- Every route from `repo/cmd/server/main.go` is represented in this document.
- Body schema references map directly to DTO types in `repo/internal/dto`.
- Response schema references map directly to model/DTO outputs returned by handlers.
- Behavioral constraints included here are taken from service/middleware/worker implementation, not inferred assumptions.

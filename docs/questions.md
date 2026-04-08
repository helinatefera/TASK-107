## 1. Password hashing algorithm

* **Question:** Passwords must be stored as salted hashes, but the hashing algorithm is not specified.
* **Assumption:** A modern, secure, adaptive hashing algorithm is required.
* **Solution:** Use Argon2id (preferred) or bcrypt with configurable cost parameters and per-password salt.

## 2. Definition of password character classes

* **Question:** The prompt requires 3 of 4 character classes but does not define them.
* **Assumption:** Classes are lowercase, uppercase, digits, and special characters.
* **Solution:** Validate passwords against these four categories.

## 3. Recovery token delivery mechanism

* **Question:** Recovery tokens are local (no email/SMS), but delivery is not defined.
* **Assumption:** Tokens are returned once via API and handled offline by the client/operator.
* **Solution:** Return token only at creation and store only hashed version.

## 4. Recovery token expiration and reuse

* **Question:** No rules for expiry or reuse are defined.
* **Assumption:** Tokens are single-use and time-limited.
* **Solution:** Store `expires_at` and `used_at`; invalidate after use or expiry.

## 5. Session refresh vs absolute expiry

* **Question:** Refresh exists, but effect on 7-day expiry is unclear.
* **Assumption:** Refresh resets idle timeout only, not absolute lifetime.
* **Solution:** Enforce hard cap of 7 days from session creation.

## 6. Device identity definition

* **Question:** Per-device sessions are required, but device identity is undefined.
* **Assumption:** Device ID is client-generated and stable.
* **Solution:** Require device_id in auth flows and bind sessions to it.

## 7. Failed login tracking scope

* **Question:** Lockout rule exists, but scope (account/IP/device) is unclear.
* **Assumption:** Lockout applies per account.
* **Solution:** Track failed attempts per user account.

## 8. Pricing precedence (station vs device)

* **Question:** Both station and device pricing exist, but precedence is undefined.
* **Assumption:** Device-level overrides station-level.
* **Solution:** Resolve device pricing first, fallback to station.

## 9. Rollback with monotonic versioning

* **Question:** Rollback is required but version numbers must be monotonic.
* **Assumption:** Rollback creates a new version instead of reusing old.
* **Solution:** Clone previous version into a new version with incremented number.

## 10. Duration rounding rule

* **Question:** Duration pricing exists, but rounding is undefined.
* **Assumption:** Partial minutes are rounded up.
* **Solution:** Apply ceiling rounding to minutes.

## 11. Sales tax rate source

* **Question:** Tax flag exists, but rate source is not defined.
* **Assumption:** Rate comes from local configuration.
* **Solution:** Resolve tax rate from config and store in snapshot.

## 12. Timezone for TOU and quiet hours

* **Question:** “Local time” is referenced but not defined.
* **Assumption:** Organization (or station) defines timezone.
* **Solution:** Store timezone and compute logic in that context.

## 13. Offline “delivered-to-inbox” definition

* **Question:** Delivery metric exists without push system.
* **Assumption:** Delivery means persisted and visible in inbox.
* **Solution:** Mark delivered when stored and retrievable.

## 14. Sensitive data masking scope

* **Question:** Masking required, but who sees unmasked data is unclear.
* **Assumption:** Only administrators can see full values.
* **Solution:** Apply role-based masking at response level.

## 15. Background job execution (offline system)

* **Question:** System must be offline but requires scheduling (notifications, etc.).
* **Assumption:** No external queue systems are allowed.
* **Solution:** Use in-process workers with PostgreSQL-backed job tables.

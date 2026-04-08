package service

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/db"
	"github.com/chargeops/api/internal/dto"
	"github.com/chargeops/api/internal/model"
	"github.com/chargeops/api/internal/repo"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

// ---------------------------------------------------------------------------
// Price Templates
// ---------------------------------------------------------------------------

func CreatePriceTemplate(ctx context.Context, dbConn sqlx.ExtContext, orgID uuid.UUID, req *dto.CreatePriceTemplateRequest) (*model.PriceTemplate, error) {
	// Exactly one of station_id or device_id must be set.
	hasStation := req.StationID != nil
	hasDevice := req.DeviceID != nil
	if hasStation == hasDevice {
		return nil, apperror.New(400, "exactly one of station_id or device_id must be provided")
	}

	now := time.Now().UTC()
	t := &model.PriceTemplate{
		ID:        uuid.New(),
		OrgID:     orgID,
		Name:      req.Name,
		StationID: req.StationID,
		DeviceID:  req.DeviceID,
		CreatedAt: now,
	}

	if err := repo.CreateTemplate(ctx, dbConn, t); err != nil {
		return nil, err
	}
	return t, nil
}

func GetPriceTemplate(ctx context.Context, dbConn sqlx.ExtContext, id uuid.UUID) (*model.PriceTemplate, error) {
	return repo.GetTemplate(ctx, dbConn, id)
}

func ListPriceTemplates(ctx context.Context, dbConn sqlx.ExtContext, orgID *uuid.UUID, limit, offset int) ([]model.PriceTemplate, error) {
	return repo.ListTemplates(ctx, dbConn, orgID, limit, offset)
}

// ---------------------------------------------------------------------------
// Versions
// ---------------------------------------------------------------------------

func CreateVersion(ctx context.Context, database *sqlx.DB, templateID uuid.UUID, req *dto.CreateVersionRequest) (*model.PriceTemplateVersion, error) {
	// Validate non-negative rates/fees
	if req.EnergyRate.IsNegative() {
		return nil, apperror.New(400, "energy_rate must not be negative")
	}
	if req.DurationRate.IsNegative() {
		return nil, apperror.New(400, "duration_rate must not be negative")
	}
	if req.ServiceFee.IsNegative() {
		return nil, apperror.New(400, "service_fee must not be negative")
	}

	var version *model.PriceTemplateVersion

	err := db.WithTx(ctx, database, func(tx *sqlx.Tx) error {
		// 1. Get max version number for this template.
		maxNum, err := repo.GetMaxVersionNumber(ctx, tx, templateID)
		if err != nil {
			return err
		}

		// 2. If apply_tax, resolve tax_rate from app_config.
		var taxRate *decimal.Decimal
		if req.ApplyTax {
			cfg, err := repo.GetConfig(ctx, tx, "tax_rate")
			if err != nil {
				return err
			}
			parsed, err := decimal.NewFromString(cfg.Value)
			if err != nil {
				return apperror.New(500, "invalid tax_rate in app_config")
			}
			taxRate = &parsed
		}

		// 3. Insert new version.
		now := time.Now().UTC()
		version = &model.PriceTemplateVersion{
			ID:            uuid.New(),
			TemplateID:    templateID,
			VersionNumber: maxNum + 1,
			EnergyRate:    req.EnergyRate,
			DurationRate:  req.DurationRate,
			ServiceFee:    req.ServiceFee,
			ApplyTax:      req.ApplyTax,
			TaxRate:       taxRate,
			Status:        "draft",
			CreatedAt:     now,
		}

		return repo.CreateVersion(ctx, tx, version)
	})
	if err != nil {
		return nil, err
	}
	return version, nil
}

func GetVersion(ctx context.Context, dbConn sqlx.ExtContext, id uuid.UUID) (*model.PriceTemplateVersion, error) {
	return repo.GetVersion(ctx, dbConn, id)
}

func ListVersions(ctx context.Context, dbConn sqlx.ExtContext, templateID uuid.UUID) ([]model.PriceTemplateVersion, error) {
	return repo.ListVersions(ctx, dbConn, templateID)
}

func ActivateVersion(ctx context.Context, database *sqlx.DB, versionID uuid.UUID, effectiveAt *time.Time) (*model.PriceTemplateVersion, error) {
	// Validate effective_at is not in the past when explicitly provided.
	if effectiveAt != nil && effectiveAt.Before(time.Now().UTC()) {
		return nil, apperror.New(400, "effective_at must not be in the past")
	}

	var version *model.PriceTemplateVersion

	err := db.WithTx(ctx, database, func(tx *sqlx.Tx) error {
		// 1. Get the version to activate.
		v, err := repo.GetVersion(ctx, tx, versionID)
		if err != nil {
			return err
		}

		// 2. Determine effective time.
		now := time.Now().UTC()
		eff := now
		if effectiveAt != nil {
			eff = effectiveAt.UTC()
		}

		// 3. Only deactivate the current active version if the new version
		// takes effect immediately. When effective_at is in the future, the
		// old version must stay active to avoid a pricing gap — the
		// resolution query (ORDER BY effective_at DESC LIMIT 1) will
		// naturally prefer the newer version once its effective_at arrives.
		if !eff.After(now) {
			if err := repo.DeactivateCurrentActive(ctx, tx, v.TemplateID); err != nil {
				return err
			}
		}

		// 4. Set this version to active with the chosen effective_at.
		if err := repo.ActivateVersion(ctx, tx, versionID, eff); err != nil {
			return err
		}

		v.Status = "active"
		v.EffectiveAt = &eff
		version = v
		return nil
	})
	if err != nil {
		return nil, err
	}
	return version, nil
}

func DeactivateVersion(ctx context.Context, dbConn sqlx.ExtContext, versionID uuid.UUID) error {
	return repo.DeactivateVersion(ctx, dbConn, versionID)
}

func RollbackVersion(ctx context.Context, database *sqlx.DB, versionID uuid.UUID) (*model.PriceTemplateVersion, error) {
	var newVersion *model.PriceTemplateVersion

	err := db.WithTx(ctx, database, func(tx *sqlx.Tx) error {
		// 1. Load target version.
		target, err := repo.GetVersion(ctx, tx, versionID)
		if err != nil {
			return err
		}

		// Load its TOU rules.
		rules, err := repo.ListTOURules(ctx, tx, versionID)
		if err != nil {
			return err
		}

		// 2. Get max version number for the template.
		maxNum, err := repo.GetMaxVersionNumber(ctx, tx, target.TemplateID)
		if err != nil {
			return err
		}

		// 3. Create a NEW version cloning all data.
		now := time.Now().UTC()
		clonedFromID := target.ID
		newVersion = &model.PriceTemplateVersion{
			ID:                  uuid.New(),
			TemplateID:          target.TemplateID,
			VersionNumber:       maxNum + 1,
			EnergyRate:          target.EnergyRate,
			DurationRate:        target.DurationRate,
			ServiceFee:          target.ServiceFee,
			ApplyTax:            target.ApplyTax,
			TaxRate:             target.TaxRate,
			Status:              "draft",
			ClonedFromVersionID: &clonedFromID,
			CreatedAt:           now,
		}
		if err := repo.CreateVersion(ctx, tx, newVersion); err != nil {
			return err
		}

		// 4. Clone all TOU rules to the new version.
		for _, r := range rules {
			cloned := &model.TOURule{
				ID:           uuid.New(),
				VersionID:    newVersion.ID,
				DayType:      r.DayType,
				StartTime:    NormalizeTimeStr(r.StartTime),
				EndTime:      NormalizeTimeStr(r.EndTime),
				EnergyRate:   r.EnergyRate,
				DurationRate: r.DurationRate,
			}
			if err := repo.CreateTOURule(ctx, tx, cloned); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return newVersion, nil
}

// ---------------------------------------------------------------------------
// TOU Rules
// ---------------------------------------------------------------------------

func CreateTOURule(ctx context.Context, dbConn sqlx.ExtContext, versionID uuid.UUID, req *dto.CreateTOURuleRequest) (*model.TOURule, error) {
	// 1. Parse and validate times.
	// Validate non-negative rates
	if req.EnergyRate.IsNegative() {
		return nil, apperror.New(400, "energy_rate must not be negative")
	}
	if req.DurationRate.IsNegative() {
		return nil, apperror.New(400, "duration_rate must not be negative")
	}

	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		return nil, apperror.ErrInvalidTimeRange
	}
	endTime, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		return nil, apperror.ErrInvalidTimeRange
	}
	if !startTime.Before(endTime) {
		return nil, apperror.ErrInvalidTimeRange
	}

	// 2. Get existing TOU rules for this version.
	existing, err := repo.ListTOURules(ctx, dbConn, versionID)
	if err != nil {
		return nil, err
	}

	// 3. Check for overlapping windows within same day_type.
	if CheckTOUOverlap(existing, req.DayType, req.StartTime, req.EndTime) {
		return nil, apperror.ErrTOUOverlap
	}

	// 5. Insert rule.
	r := &model.TOURule{
		ID:           uuid.New(),
		VersionID:    versionID,
		DayType:      req.DayType,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		EnergyRate:   req.EnergyRate,
		DurationRate: req.DurationRate,
	}
	if err := repo.CreateTOURule(ctx, dbConn, r); err != nil {
		return nil, err
	}
	return r, nil
}

func ListTOURules(ctx context.Context, dbConn sqlx.ExtContext, versionID uuid.UUID) ([]model.TOURule, error) {
	return repo.ListTOURules(ctx, dbConn, versionID)
}

func GetTOURule(ctx context.Context, dbConn sqlx.ExtContext, id uuid.UUID) (*model.TOURule, error) {
	return repo.GetTOURule(ctx, dbConn, id)
}

func DeleteTOURule(ctx context.Context, dbConn sqlx.ExtContext, id uuid.UUID) error {
	return repo.DeleteTOURule(ctx, dbConn, id)
}

// ---------------------------------------------------------------------------
// Order Snapshots
// ---------------------------------------------------------------------------

func CreateOrderSnapshot(ctx context.Context, database *sqlx.DB, userID uuid.UUID, req *dto.CreateOrderRequest, callerOrgID *uuid.UUID, isAdmin bool) (*model.OrderSnapshot, error) {
	var snapshot *model.OrderSnapshot

	err := db.WithTx(ctx, database, func(tx *sqlx.Tx) error {
		// 1. Get device to find station_id.
		device, err := repo.GetDevice(ctx, tx, req.DeviceID)
		if err != nil {
			return err
		}

		// 1b. Tenant guard: non-admin callers must have an org matching the
		// device's station, or the station's org must be a child of the caller's org.
		if !isAdmin {
			if callerOrgID == nil {
				return apperror.ErrForbidden
			}
			station, err := repo.GetStation(ctx, tx, device.StationID)
			if err != nil {
				return err
			}
			if station.OrgID != *callerOrgID {
				accessible, aErr := repo.IsOrgAccessible(ctx, tx, *callerOrgID, station.OrgID)
				if aErr != nil || !accessible {
					return apperror.ErrForbidden
				}
			}
		}

		// 2. Resolve pricing version.
		version, err := resolvePricing(ctx, tx, req.DeviceID, req.StartTime)
		if err != nil {
			return err
		}

		// 3. Get station timezone for TOU matching.
		station, err := repo.GetStation(ctx, tx, device.StationID)
		if err != nil {
			return err
		}

		loc, err := time.LoadLocation(station.Timezone)
		if err != nil {
			return apperror.New(500, "invalid station timezone: "+station.Timezone)
		}

		// 4. Check TOU rules.
		rules, err := repo.ListTOURules(ctx, tx, version.ID)
		if err != nil {
			return err
		}

		localStart := req.StartTime.In(loc)
		dayType := ClassifyDayType(localStart)
		touRule := FindMatchingTOURule(rules, dayType, localStart)
		// Fall back to holiday rules if no weekday/weekend rule matched
		if touRule == nil {
			touRule = FindMatchingTOURule(rules, "holiday", localStart)
		}

		// Determine effective rates.
		energyRate := version.EnergyRate
		durationRate := version.DurationRate
		var touApplied json.RawMessage
		if touRule != nil {
			energyRate = touRule.EnergyRate
			durationRate = touRule.DurationRate
			touJSON, err := json.Marshal(touRule)
			if err != nil {
				return err
			}
			touApplied = touJSON
		}

		// 5. Calculate costs.
		durationSec := req.EndTime.Sub(req.StartTime).Seconds()
		durationMin := int(math.Ceil(durationSec / 60.0))

		energyCost := energyRate.Mul(req.EnergyKWh)
		durationCost := durationRate.Mul(decimal.NewFromInt(int64(durationMin)))
		subtotal := energyCost.Add(durationCost).Add(version.ServiceFee)

		taxAmount := decimal.Zero
		if version.ApplyTax && version.TaxRate != nil {
			taxAmount = subtotal.Mul(*version.TaxRate)
		}
		total := subtotal.Add(taxAmount)

		now := time.Now().UTC()
		snapshot = &model.OrderSnapshot{
			ID:           uuid.New(),
			OrderID:      req.OrderID,
			UserID:       userID,
			DeviceID:     req.DeviceID,
			StationID:    device.StationID,
			VersionID:    version.ID,
			EnergyRate:   energyRate,
			DurationRate: durationRate,
			ServiceFee:   version.ServiceFee,
			TaxRate:      version.TaxRate,
			TOUApplied:   touApplied,
			EnergyKWh:    req.EnergyKWh,
			DurationMin:  durationMin,
			Subtotal:     subtotal,
			TaxAmount:    taxAmount,
			Total:        total,
			OrderStart:   req.StartTime,
			OrderEnd:     req.EndTime,
			CreatedAt:    now,
		}

		return repo.CreateOrderSnapshot(ctx, tx, snapshot)
	})
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

func GetOrderSnapshot(ctx context.Context, dbConn sqlx.ExtContext, id uuid.UUID) (*model.OrderSnapshot, error) {
	return repo.GetOrderSnapshot(ctx, dbConn, id)
}

func ListOrderSnapshots(ctx context.Context, dbConn sqlx.ExtContext, userID *uuid.UUID, limit, offset int) ([]model.OrderSnapshot, error) {
	return repo.ListOrderSnapshots(ctx, dbConn, userID, limit, offset)
}

func RecalculateOrder(ctx context.Context, database *sqlx.DB, orderID uuid.UUID) (*model.OrderSnapshot, error) {
	// 1. Get existing order snapshot.
	existing, err := repo.GetOrderSnapshot(ctx, database, orderID)
	if err != nil {
		return nil, err
	}

	var result *model.OrderSnapshot

	err = db.WithTx(ctx, database, func(tx *sqlx.Tx) error {
		// 2. Use the pricing version that was active at the order's original start time.
		// This is the version stored on the order snapshot — it reflects the historical
		// pricing regardless of whether the version has since been deactivated.
		version, err := repo.GetVersion(ctx, tx, existing.VersionID)
		if err != nil {
			return err
		}

		// Get station timezone.
		station, err := repo.GetStation(ctx, tx, existing.StationID)
		if err != nil {
			return err
		}

		loc, err := time.LoadLocation(station.Timezone)
		if err != nil {
			return apperror.New(500, "invalid station timezone: "+station.Timezone)
		}

		// Check TOU rules.
		rules, err := repo.ListTOURules(ctx, tx, version.ID)
		if err != nil {
			return err
		}

		localStart := existing.OrderStart.In(loc)
		dayType := ClassifyDayType(localStart)
		touRule := FindMatchingTOURule(rules, dayType, localStart)

		energyRate := version.EnergyRate
		durationRate := version.DurationRate
		var touApplied json.RawMessage
		if touRule != nil {
			energyRate = touRule.EnergyRate
			durationRate = touRule.DurationRate
			touJSON, err := json.Marshal(touRule)
			if err != nil {
				return err
			}
			touApplied = touJSON
		}

		// 3. Recalculate with same energy_kwh and duration.
		durationSec := existing.OrderEnd.Sub(existing.OrderStart).Seconds()
		durationMin := int(math.Ceil(durationSec / 60.0))

		energyCost := energyRate.Mul(existing.EnergyKWh)
		durationCost := durationRate.Mul(decimal.NewFromInt(int64(durationMin)))
		subtotal := energyCost.Add(durationCost).Add(version.ServiceFee)

		taxAmount := decimal.Zero
		if version.ApplyTax && version.TaxRate != nil {
			taxAmount = subtotal.Mul(*version.TaxRate)
		}
		total := subtotal.Add(taxAmount)

		// 4. Return computed snapshot without persisting.
		result = &model.OrderSnapshot{
			ID:           existing.ID,
			OrderID:      existing.OrderID,
			UserID:       existing.UserID,
			DeviceID:     existing.DeviceID,
			StationID:    existing.StationID,
			VersionID:    version.ID,
			EnergyRate:   energyRate,
			DurationRate: durationRate,
			ServiceFee:   version.ServiceFee,
			TaxRate:      version.TaxRate,
			TOUApplied:   touApplied,
			EnergyKWh:    existing.EnergyKWh,
			DurationMin:  durationMin,
			Subtotal:     subtotal,
			TaxAmount:    taxAmount,
			Total:        total,
			OrderStart:   existing.OrderStart,
			OrderEnd:     existing.OrderEnd,
			CreatedAt:    existing.CreatedAt,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// resolvePricing tries device-level active version first, then falls back to
// station-level.
func resolvePricing(ctx context.Context, dbConn sqlx.ExtContext, deviceID uuid.UUID, at time.Time) (*model.PriceTemplateVersion, error) {
	v, err := repo.GetActiveVersionForDevice(ctx, dbConn, deviceID, at)
	if err == nil {
		return v, nil
	}
	if !errors.Is(err, apperror.ErrNotFound) {
		return nil, err
	}

	// Fallback: look up station via device.
	device, err := repo.GetDevice(ctx, dbConn, deviceID)
	if err != nil {
		return nil, err
	}

	v, err = repo.GetActiveVersionForStation(ctx, dbConn, device.StationID, at)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// ClassifyDayType returns "weekday" for Mon-Fri and "weekend" for Sat-Sun.
func ClassifyDayType(t time.Time) string {
	switch t.Weekday() {
	case time.Saturday, time.Sunday:
		return "weekend"
	default:
		return "weekday"
	}
}

// CheckTOUOverlap returns true if a new TOU window (dayType, startTime, endTime)
// overlaps any existing rule of the same day type.
func CheckTOUOverlap(existing []model.TOURule, dayType, startTime, endTime string) bool {
	nStart := NormalizeTimeStr(startTime)
	nEnd := NormalizeTimeStr(endTime)
	for _, rule := range existing {
		if rule.DayType != dayType {
			continue
		}
		rStart := NormalizeTimeStr(rule.StartTime)
		rEnd := NormalizeTimeStr(rule.EndTime)
		if nStart < rEnd && nEnd > rStart {
			return true
		}
	}
	return false
}

// FindMatchingTOURule returns the first rule whose day_type matches and whose
// time window contains localTime, or nil if no rule matches.
func FindMatchingTOURule(rules []model.TOURule, dayType string, localTime time.Time) *model.TOURule {
	target := localTime.Format("15:04")

	for i := range rules {
		if rules[i].DayType != dayType {
			continue
		}
		start := NormalizeTimeStr(rules[i].StartTime)
		end := NormalizeTimeStr(rules[i].EndTime)

		// Match when target >= start AND target < end.
		if target >= start && target < end {
			return &rules[i]
		}
	}
	return nil
}

// NormalizeTimeStr normalizes a time string to "HH:MM" for lexicographic comparison.
// Handles: "07:00", "07:00:00", "0000-01-01T07:00:00Z" (PostgreSQL TIME scanned as RFC3339).
func NormalizeTimeStr(s string) string {
	s = strings.TrimSpace(s)
	// Handle PostgreSQL TIME scanned as full timestamp "0000-01-01T07:00:00Z"
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Format("15:04")
	}
	// Handle "HH:MM:SS" -> "HH:MM"
	if len(s) > 5 && s[2] == ':' {
		return s[:5]
	}
	return s
}

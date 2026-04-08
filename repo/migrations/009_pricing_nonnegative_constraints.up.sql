-- Enforce non-negative rates and fees at the DB level
ALTER TABLE price_template_versions ADD CONSTRAINT chk_ptv_energy_rate_nonneg CHECK (energy_rate >= 0);
ALTER TABLE price_template_versions ADD CONSTRAINT chk_ptv_duration_rate_nonneg CHECK (duration_rate >= 0);
ALTER TABLE price_template_versions ADD CONSTRAINT chk_ptv_service_fee_nonneg CHECK (service_fee >= 0);
ALTER TABLE price_template_versions ADD CONSTRAINT chk_ptv_tax_rate_nonneg CHECK (tax_rate IS NULL OR tax_rate >= 0);

ALTER TABLE tou_rules ADD CONSTRAINT chk_tou_energy_rate_nonneg CHECK (energy_rate >= 0);
ALTER TABLE tou_rules ADD CONSTRAINT chk_tou_duration_rate_nonneg CHECK (duration_rate >= 0);

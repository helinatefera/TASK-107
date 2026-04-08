ALTER TABLE tou_rules DROP CONSTRAINT IF EXISTS chk_tou_duration_rate_nonneg;
ALTER TABLE tou_rules DROP CONSTRAINT IF EXISTS chk_tou_energy_rate_nonneg;

ALTER TABLE price_template_versions DROP CONSTRAINT IF EXISTS chk_ptv_tax_rate_nonneg;
ALTER TABLE price_template_versions DROP CONSTRAINT IF EXISTS chk_ptv_service_fee_nonneg;
ALTER TABLE price_template_versions DROP CONSTRAINT IF EXISTS chk_ptv_duration_rate_nonneg;
ALTER TABLE price_template_versions DROP CONSTRAINT IF EXISTS chk_ptv_energy_rate_nonneg;

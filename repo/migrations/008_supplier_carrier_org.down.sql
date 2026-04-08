DROP INDEX IF EXISTS idx_carriers_dedup;
DROP INDEX IF EXISTS idx_suppliers_dedup;

ALTER TABLE carriers DROP COLUMN IF EXISTS org_id;
ALTER TABLE suppliers DROP COLUMN IF EXISTS org_id;

CREATE UNIQUE INDEX idx_suppliers_dedup ON suppliers(normalized_name, tax_id) WHERE tax_id IS NOT NULL;
CREATE UNIQUE INDEX idx_carriers_dedup ON carriers(normalized_name, tax_id) WHERE tax_id IS NOT NULL;

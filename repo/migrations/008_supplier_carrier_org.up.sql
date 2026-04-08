ALTER TABLE suppliers ADD COLUMN org_id UUID REFERENCES organizations(id);
ALTER TABLE carriers ADD COLUMN org_id UUID REFERENCES organizations(id);

-- Replace global dedup indexes with org-scoped ones
DROP INDEX IF EXISTS idx_suppliers_dedup;
CREATE UNIQUE INDEX idx_suppliers_dedup ON suppliers(org_id, normalized_name, tax_id) WHERE tax_id IS NOT NULL AND org_id IS NOT NULL;

DROP INDEX IF EXISTS idx_carriers_dedup;
CREATE UNIQUE INDEX idx_carriers_dedup ON carriers(org_id, normalized_name, tax_id) WHERE tax_id IS NOT NULL AND org_id IS NOT NULL;

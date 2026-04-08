-- DB-level exclusion constraint to prevent TOU time overlaps within the same version+day_type.
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Create a custom range type for TIME
CREATE TYPE timerange AS RANGE (subtype = time);

ALTER TABLE tou_rules ADD CONSTRAINT excl_tou_no_overlap
    EXCLUDE USING gist (
        version_id WITH =,
        day_type WITH =,
        timerange(start_time, end_time) WITH &&
    );

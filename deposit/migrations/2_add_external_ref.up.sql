ALTER TABLE deposit
ADD COLUMN external_ref TEXT;

CREATE INDEX external_ref_index
ON deposit (external_ref);
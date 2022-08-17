ALTER TABLE organization
RENAME COLUMN pub_key TO signing_pub_key;

ALTER TABLE organization
ADD encryption_pub_key TEXT NOT NULL DEFAULT '';
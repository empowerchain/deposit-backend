CREATE TABLE deposit
(
    id                       TEXT PRIMARY KEY,
    scheme_id                TEXT      NOT NULL,
    collection_point_pub_key TEXT      NOT NULL,
    user_pub_key             TEXT      NOT NULL DEFAULT '',
    mass_balance_deposits    JSON      NOT NULL,
    claimed                  BOOL      NOT NULL DEFAULT false,
    created_at               TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE voucher_definition
(
    id              TEXT PRIMARY KEY,
    organization_id TEXT      NOT NULL,
    name            TEXT      NOT NULL,
    picture_url     TEXT      NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE voucher
(
    id                    TEXT PRIMARY KEY,
    voucher_definition_id TEXT      NOT NULL,
    owner_pub_key         TEXT,
    invalidated           BOOL      NOT NULL,
    created_at            TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT fk_voucher_definition FOREIGN KEY (voucher_definition_id) REFERENCES voucher_definition (id)
)
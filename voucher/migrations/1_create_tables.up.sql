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
    created_at            TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT fk_voucher_definition FOREIGN KEY (voucher_definition_id) REFERENCES voucher_definition (id)
)
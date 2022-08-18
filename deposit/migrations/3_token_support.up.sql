CREATE TABLE token_definition
(
    id              TEXT PRIMARY KEY,
    organization_id TEXT      NOT NULL,
    name            TEXT      NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE token_balance
(
    token_definition_id TEXT    NOT NULL,
    owner_pub_key       TEXT    NOT NULL,
    balance             DECIMAL NOT NULL DEFAULT 0.0,
    PRIMARY KEY(token_definition_id, owner_pub_key)
)
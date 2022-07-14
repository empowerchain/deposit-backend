CREATE TABLE deposit
(
    id                       TEXT PRIMARY KEY,
    scheme_id                TEXT      NOT NULL,
    collection_point_pub_key TEXT      NOT NULL,
    user_pub_key             TEXT,
    mass_balance_deposits    JSON      NOT NULL,
    created_at               TIMESTAMP NOT NULL DEFAULT now()
);
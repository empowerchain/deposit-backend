CREATE TABLE user_organization
(
    id                    TEXT PRIMARY KEY,
    organization_pub_key  TEXT      NOT NULL,
    user_pub_key          TEXT      NOT NULL,
    content               TEXT      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);
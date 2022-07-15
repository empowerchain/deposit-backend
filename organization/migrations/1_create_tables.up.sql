CREATE TABLE organization
(
    id         TEXT PRIMARY KEY,
    name       TEXT      NOT NULL,
    pub_key    TEXT      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);
CREATE TABLE scheme
(
    id         TEXT PRIMARY KEY,
    name       TEXT      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);
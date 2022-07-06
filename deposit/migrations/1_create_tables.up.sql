CREATE TABLE scheme
(
    id         TEXT PRIMARY KEY,
    name       TEXT      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE deposit
(
    id                  TEXT PRIMARY KEY,
    scheme_id           TEXT      NOT NULL,
    collection_point_id TEXT      NOT NULL,
    user_id             TEXT,
    created_at          TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT fk_scheme FOREIGN KEY (scheme_id) REFERENCES scheme (id)
);
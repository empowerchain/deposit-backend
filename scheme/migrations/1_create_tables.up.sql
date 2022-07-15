CREATE TABLE scheme
(
    id                 TEXT PRIMARY KEY,
    organization_id    TEXT      NOT NULL,
    name               TEXT      NOT NULL,
    collection_points  TEXT[],
    reward_definitions JSON      NOT NULL,
    created_at         TIMESTAMP NOT NULL DEFAULT now()
);
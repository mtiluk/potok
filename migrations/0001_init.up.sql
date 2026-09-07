CREATE TABLE users (
    id            TEXT      PRIMARY KEY,
    email         TEXT      NOT NULL UNIQUE,
    password_hash TEXT      NOT NULL,
    is_admin      INTEGER   NOT NULL DEFAULT 0,
    api_key       TEXT      UNIQUE,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE vaults (
    id          TEXT      PRIMARY KEY,
    user_id     TEXT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name        TEXT      NOT NULL,
    generation  INTEGER   NOT NULL DEFAULT 0,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, name)
);

CREATE INDEX vaults_user_id_idx ON vaults (user_id);

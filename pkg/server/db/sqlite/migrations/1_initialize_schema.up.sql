-- create tables
CREATE TABLE IF NOT EXISTS trust_domains
(
    id          text PRIMARY KEY,
    name        TEXT      NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS relationships
(
    id                     TEXT PRIMARY KEY,
    trust_domain_a_id      TEXT      NOT NULL,
    trust_domain_b_id      TEXT      NOT NULL,
    trust_domain_a_consent TEXT      NOT NULL DEFAULT 'pending',
    trust_domain_b_consent TEXT      NOT NULL DEFAULT 'pending',
    created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (trust_domain_a_id, trust_domain_b_id),
    FOREIGN KEY (trust_domain_a_id)
        REFERENCES trust_domains (id),
    FOREIGN KEY (trust_domain_b_id)
        REFERENCES trust_domains (id)
);

CREATE TABLE IF NOT EXISTS bundles
(
    id                  TEXT PRIMARY KEY,
    trust_domain_id     TEXT      NOT NULL UNIQUE,
    data                BLOB      NOT NULL,
    digest              BLOB      NOT NULL,
    signature           BLOB,
    signing_certificate BLOB,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (trust_domain_id)
        REFERENCES trust_domains (id)
);

CREATE TABLE IF NOT EXISTS join_tokens
(
    id              TEXT PRIMARY KEY,
    trust_domain_id TEXT      NOT NULL,
    token           TEXT      NOT NULL UNIQUE,
    used            BOOL      NOT NULL DEFAULT 0,
    expires_at      TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (trust_domain_id)
        REFERENCES trust_domains (id)
);

-- create tables

CREATE TYPE consent_status AS ENUM ('approved', 'denied', 'pending');

CREATE TABLE IF NOT EXISTS trust_domains
(
    id          UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    name        TEXT                     NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS relationships
(
    id                     UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    trust_domain_a_id      UUID                     NOT NULL,
    trust_domain_b_id      UUID                     NOT NULL,
    trust_domain_a_consent consent_status           NOT NULL DEFAULT 'pending',
    trust_domain_b_consent consent_status           NOT NULL DEFAULT 'pending',
    created_at             TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at             TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    UNIQUE (trust_domain_a_id, trust_domain_b_id)
);

CREATE TABLE IF NOT EXISTS bundles
(
    id                  UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    trust_domain_id     UUID                     NOT NULL UNIQUE,
    data                BYTEA                    NOT NULL,
    digest              BYTEA                    NOT NULL,
    signature           BYTEA,
    signing_certificate BYTEA,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS join_tokens
(
    id              UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    trust_domain_id UUID                     NOT NULL,
    token           TEXT                     NOT NULL UNIQUE,
    used            BOOL                     NOT NULL DEFAULT FALSE,
    expires_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- define foreign keys
ALTER TABLE "relationships"
    ADD FOREIGN KEY ("trust_domain_a_id") REFERENCES "trust_domains" ("id");
ALTER TABLE "relationships"
    ADD FOREIGN KEY ("trust_domain_b_id") REFERENCES "trust_domains" ("id");

ALTER TABLE "bundles"
    ADD FOREIGN KEY ("trust_domain_id") REFERENCES "trust_domains" ("id");

ALTER TABLE "join_tokens"
    ADD FOREIGN KEY ("trust_domain_id") REFERENCES "trust_domains" ("id");

-- create indexes
-- PostgresSQL automatically creates a unique index when a unique constraint or primary key is defined for a table
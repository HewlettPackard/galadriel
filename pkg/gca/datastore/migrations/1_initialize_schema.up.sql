-- Migration version 1. 
-- If changes on tables are needed, a new version file should be created with the version number as a prefix
-- That prefix should match the currentDBVersion in schemas.go

CREATE TABLE IF NOT EXISTS trust_domains
(
    id                  UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name                TEXT                     NOT NULL UNIQUE,
    description         TEXT,
    harvester_spiffe_id TEXT,
    onboarding_bundle   BYTEA,
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

ALTER TABLE "join_tokens"
    ADD FOREIGN KEY ("trust_domain_id") REFERENCES "trust_domains" ("id");

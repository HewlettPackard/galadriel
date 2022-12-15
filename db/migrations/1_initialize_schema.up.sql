-- create tables

CREATE TYPE status AS ENUM ('pending', 'active', 'disabled', 'denied');

CREATE TABLE federation_groups
(
    id          UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    name        TEXT                     NOT NULL UNIQUE,
    description TEXT,
    status      status                   NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE TABLE members
(
    id           UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    trust_domain TEXT                     NOT NULL UNIQUE,
    status       status                   NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE TABLE memberships
(
    id                  UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    member_id           UUID                     NOT NULL,
    federation_group_id UUID                     NOT NULL,
    status              status                   NOT NULL,
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE TABLE bundles
(
    id            UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    raw_bundle    BYTEA                    NOT NULL,
    digest        BYTEA                    NOT NULL,
    signed_bundle BYTEA,
    tlog_id       UUID,
    svid_pem      TEXT,
    member_id     UUID                     NOT NULL UNIQUE,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE TABLE join_tokens
(
    id         UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    token      TEXT                     NOT NULL UNIQUE,
    used       BOOL                              DEFAULT FALSE,
    expiry     TIMESTAMP WITH TIME ZONE NOT NULL,
    member_id  UUID                     NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE TABLE harvesters
(
    id           UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    member_id    UUID                     NOT NULL,
    is_leader    BOOL                              DEFAULT FALSE,
    leader_until TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- define foreign keys
ALTER TABLE "memberships"
    ADD FOREIGN KEY ("member_id") REFERENCES "members" ("id");
ALTER TABLE "memberships"
    ADD FOREIGN KEY ("federation_group_id") REFERENCES "federation_groups" ("id");

ALTER TABLE "bundles"
    ADD FOREIGN KEY ("member_id") REFERENCES "members" ("id");

ALTER TABLE "join_tokens"
    ADD FOREIGN KEY ("member_id") REFERENCES "members" ("id");

ALTER TABLE "harvesters"
    ADD FOREIGN KEY ("member_id") REFERENCES "members" ("id");

-- create indexes
-- PostgreSQL automatically creates a unique index when a unique constraint or primary key is defined for a table
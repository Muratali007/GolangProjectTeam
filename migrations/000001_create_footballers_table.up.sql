CREATE TABLE IF NOT EXISTS footballers (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    names text NOT NULL,
    titles integer NOT NULL,
    startedPlayYear integer NOT NULL,
    year integer NOT NULL,
    club text NOT NULL,
    playedClubs integer NOT NULL,
    positions text[] NOT NULL,
    goals integer NOT NULL,
    version integer NOT NULL DEFAULT 1
    );

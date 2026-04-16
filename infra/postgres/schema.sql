-- DartScheduler PostgreSQL schema
-- Run this once on a fresh database; the Go Open() function applies it automatically.

CREATE TABLE IF NOT EXISTS applied_migrations (
    name       TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS schedules (
    id               UUID PRIMARY KEY,
    competition_name TEXT        NOT NULL,
    season           TEXT        NOT NULL DEFAULT '',
    active           BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL,
    player_list_id   UUID
);

CREATE TABLE IF NOT EXISTS evenings (
    id              UUID    PRIMARY KEY,
    schedule_id     UUID    NOT NULL REFERENCES schedules(id),
    number          INTEGER NOT NULL,
    date            TIMESTAMPTZ NOT NULL,
    is_inhaal_avond BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS player_lists (
    id         UUID PRIMARY KEY,
    name       TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS players (
    id           UUID PRIMARY KEY,
    schedule_id  UUID REFERENCES schedules(id),
    list_id      UUID REFERENCES player_lists(id),
    nr           TEXT NOT NULL DEFAULT '',
    name         TEXT NOT NULL,
    email        TEXT NOT NULL DEFAULT '',
    sponsor      TEXT NOT NULL DEFAULT '',
    address      TEXT NOT NULL DEFAULT '',
    postal_code  TEXT NOT NULL DEFAULT '',
    city         TEXT NOT NULL DEFAULT '',
    phone        TEXT NOT NULL DEFAULT '',
    mobile       TEXT NOT NULL DEFAULT '',
    member_since TEXT NOT NULL DEFAULT '',
    class        TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS buddy_preferences (
    player_id UUID NOT NULL,
    buddy_id  UUID NOT NULL,
    PRIMARY KEY (player_id, buddy_id)
);

CREATE TABLE IF NOT EXISTS matches (
    id                      UUID PRIMARY KEY,
    evening_id              UUID    NOT NULL REFERENCES evenings(id),
    player_a                UUID    NOT NULL,
    player_b                UUID    NOT NULL,
    score_a                 INTEGER,
    score_b                 INTEGER,
    played                  BOOLEAN NOT NULL DEFAULT FALSE,
    leg1_winner             TEXT    NOT NULL DEFAULT '',
    leg1_turns              INTEGER NOT NULL DEFAULT 0,
    leg2_winner             TEXT    NOT NULL DEFAULT '',
    leg2_turns              INTEGER NOT NULL DEFAULT 0,
    leg3_winner             TEXT    NOT NULL DEFAULT '',
    leg3_turns              INTEGER NOT NULL DEFAULT 0,
    reported_by             TEXT    NOT NULL DEFAULT '',
    reschedule_date         TEXT    NOT NULL DEFAULT '',
    secretary_nr            TEXT    NOT NULL DEFAULT '',
    counter_nr              TEXT    NOT NULL DEFAULT '',
    player_a_180s           INTEGER NOT NULL DEFAULT 0,
    player_b_180s           INTEGER NOT NULL DEFAULT 0,
    player_a_highest_finish INTEGER NOT NULL DEFAULT 0,
    player_b_highest_finish INTEGER NOT NULL DEFAULT 0,
    played_date             TEXT    NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS season_player_stats (
    schedule_id    UUID    NOT NULL REFERENCES schedules(id),
    player_id      UUID    NOT NULL,
    one_eighties   INTEGER NOT NULL DEFAULT 0,
    highest_finish INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (schedule_id, player_id)
);

CREATE TABLE IF NOT EXISTS evening_player_stats (
    evening_id     UUID    NOT NULL REFERENCES evenings(id),
    player_id      UUID    NOT NULL,
    one_eighties   INTEGER NOT NULL DEFAULT 0,
    highest_finish INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (evening_id, player_id)
);

-- Add FK on schedules.player_list_id after player_lists exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'schedules_player_list_id_fkey'
    ) THEN
        ALTER TABLE schedules
            ADD CONSTRAINT schedules_player_list_id_fkey
            FOREIGN KEY (player_list_id) REFERENCES player_lists(id);
    END IF;
END
$$;

CREATE INDEX IF NOT EXISTS idx_matches_evening   ON matches(evening_id);
CREATE INDEX IF NOT EXISTS idx_matches_player_a  ON matches(player_a);
CREATE INDEX IF NOT EXISTS idx_matches_player_b  ON matches(player_b);
CREATE INDEX IF NOT EXISTS idx_evenings_schedule ON evenings(schedule_id);
CREATE INDEX IF NOT EXISTS idx_players_schedule  ON players(schedule_id);
CREATE INDEX IF NOT EXISTS idx_players_list      ON players(list_id);

PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS schedules (
    id               TEXT PRIMARY KEY,
    competition_name TEXT NOT NULL,
    created_at       DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS evenings (
    id          TEXT PRIMARY KEY,
    schedule_id TEXT NOT NULL REFERENCES schedules(id),
    number      INTEGER NOT NULL,
    date        DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS players (
    id      TEXT PRIMARY KEY,
    name    TEXT NOT NULL,
    email   TEXT NOT NULL DEFAULT '',
    sponsor TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS buddy_preferences (
    player_id TEXT NOT NULL,
    buddy_id  TEXT NOT NULL,
    PRIMARY KEY (player_id, buddy_id)
);

CREATE TABLE IF NOT EXISTS matches (
    id         TEXT PRIMARY KEY,
    evening_id TEXT NOT NULL REFERENCES evenings(id),
    player_a   TEXT NOT NULL,
    player_b   TEXT NOT NULL,
    score_a    INTEGER,
    score_b    INTEGER,
    played     INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_matches_evening   ON matches(evening_id);
CREATE INDEX IF NOT EXISTS idx_matches_player_a  ON matches(player_a);
CREATE INDEX IF NOT EXISTS idx_matches_player_b  ON matches(player_b);
CREATE INDEX IF NOT EXISTS idx_evenings_schedule ON evenings(schedule_id);

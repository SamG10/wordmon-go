CREATE TABLE players (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    xp INT NOT NULL DEFAULT 0,
    level INT NOT NULL DEFAULT 1
);

CREATE TABLE words (
    id TEXT PRIMARY KEY,
    text TEXT NOT NULL,
    rarity TEXT NOT NULL,
    points INT NOT NULL
);

CREATE TABLE captures (
    id UUID PRIMARY KEY,
    player_id UUID REFERENCES players (id) ON DELETE CASCADE,
    word_id TEXT REFERENCES words (id) ON DELETE CASCADE,
    captured_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
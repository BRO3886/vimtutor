package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS sessions (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	lesson_id   TEXT NOT NULL,
	mode        TEXT NOT NULL,
	started_at  TEXT NOT NULL,
	ended_at    TEXT,
	created_at  TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS keystroke_events (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	session_id   INTEGER NOT NULL REFERENCES sessions(id),
	ts           TEXT NOT NULL,
	key          TEXT NOT NULL,
	was_correct  INTEGER NOT NULL DEFAULT 0,
	challenge_id TEXT
);
CREATE INDEX IF NOT EXISTS idx_ke_session ON keystroke_events(session_id);

CREATE TABLE IF NOT EXISTS lesson_results (
	id              INTEGER PRIMARY KEY AUTOINCREMENT,
	session_id      INTEGER NOT NULL REFERENCES sessions(id),
	lesson_id       TEXT NOT NULL,
	step_id         TEXT NOT NULL,
	attempts        INTEGER NOT NULL DEFAULT 0,
	keystrokes_used INTEGER NOT NULL DEFAULT 0,
	time_spent_secs REAL NOT NULL DEFAULT 0,
	passed          INTEGER NOT NULL DEFAULT 0,
	mistake_keys    TEXT,
	completed_at    TEXT DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_lr_lesson ON lesson_results(lesson_id);

CREATE TABLE IF NOT EXISTS daily_stats (
	date                TEXT PRIMARY KEY,
	total_keystrokes    INTEGER DEFAULT 0,
	correct_keystrokes  INTEGER DEFAULT 0,
	time_spent_secs     REAL DEFAULT 0,
	lessons_completed   INTEGER DEFAULT 0,
	xp_earned           INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS user_progress (
	lesson_id       TEXT PRIMARY KEY,
	best_keystrokes INTEGER,
	best_time_secs  REAL,
	completed_count INTEGER DEFAULT 0,
	last_attempted  TEXT,
	unlocked        INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS user_xp (
	id          INTEGER PRIMARY KEY CHECK (id = 1),
	total_xp    INTEGER DEFAULT 0
);
INSERT OR IGNORE INTO user_xp (id, total_xp) VALUES (1, 0);
`

// DB wraps the SQLite database connection.
type DB struct {
	db *sql.DB
}

// Open opens (or creates) the vimtutor database.
func Open() (*DB, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("find home dir: %w", err)
	}
	dir := filepath.Join(home, ".vimtutor")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	path := filepath.Join(dir, "vimtutor.db")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &DB{db: db}, nil
}

// Close closes the database.
func (d *DB) Close() error {
	return d.db.Close()
}

// DB returns the underlying *sql.DB for queries.
func (d *DB) DB() *sql.DB {
	return d.db
}

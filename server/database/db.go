package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func Init(dbPath string) *sql.DB {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	migrate(db)
	return db
}

func migrate(db *sql.DB) {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS family (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			name        TEXT NOT NULL,
			description TEXT,
			webhook_key TEXT DEFAULT '',     -- 企业微信 webhook key，为空则不推送
			created_at  DATETIME DEFAULT (datetime('now','localtime')),
			updated_at  DATETIME DEFAULT (datetime('now','localtime'))
		)`,
		`CREATE TABLE IF NOT EXISTS person (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			family_id   INTEGER NOT NULL,
			name        TEXT NOT NULL,
			gender      TEXT CHECK(gender IN ('male','female','unknown')) DEFAULT 'unknown',
			birthday    TEXT,
			birthday_type TEXT CHECK(birthday_type IN ('solar','lunar')) DEFAULT 'solar',
			generation  INTEGER,
			photo_url   TEXT,
			phone       TEXT,
			address     TEXT,
			notes       TEXT,
			is_alive    INTEGER DEFAULT 1,
			created_at  DATETIME DEFAULT (datetime('now','localtime')),
			updated_at  DATETIME DEFAULT (datetime('now','localtime')),
			FOREIGN KEY (family_id) REFERENCES family(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS relation (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			person_id   INTEGER NOT NULL,
			related_id  INTEGER NOT NULL,
			type        TEXT NOT NULL CHECK(type IN ('parent','child','spouse','sibling','in_law','grandparent','grandchild')),
			created_at  DATETIME DEFAULT (datetime('now','localtime')),
			FOREIGN KEY (person_id) REFERENCES person(id) ON DELETE CASCADE,
			FOREIGN KEY (related_id) REFERENCES person(id) ON DELETE CASCADE,
			UNIQUE(person_id, related_id, type)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_person_family ON person(family_id)`,
		`CREATE INDEX IF NOT EXISTS idx_relation_person ON relation(person_id)`,
		`CREATE INDEX IF NOT EXISTS idx_relation_related ON relation(related_id)`,
		`CREATE INDEX IF NOT EXISTS idx_person_birthday ON person(birthday)`,
		`ALTER TABLE person ADD COLUMN birthday_type TEXT CHECK(birthday_type IN ('solar','lunar')) DEFAULT 'solar'`,
	}

	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			log.Fatalf("migration failed: %v\nSQL: %s", err, s)
		}
	}

	log.Println("database migration complete")
}

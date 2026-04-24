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

func hasColumn(db *sql.DB, table, col string) bool {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		log.Fatalf("migration failed: pragma table_info: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			log.Fatalf("migration failed: scan table_info: %v", err)
		}
		if name == col {
			return true
		}
	}
	return false
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
			birth_date  TEXT,
			death_date  TEXT,
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
			type        TEXT NOT NULL CHECK(type IN ('parent','child','spouse')),
			created_at  DATETIME DEFAULT (datetime('now','localtime')),
			FOREIGN KEY (person_id) REFERENCES person(id) ON DELETE CASCADE,
			FOREIGN KEY (related_id) REFERENCES person(id) ON DELETE CASCADE,
			UNIQUE(person_id, related_id, type)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_person_family ON person(family_id)`,
		`CREATE INDEX IF NOT EXISTS idx_relation_person ON relation(person_id)`,
		`CREATE INDEX IF NOT EXISTS idx_relation_related ON relation(related_id)`,
		`CREATE INDEX IF NOT EXISTS idx_person_birthday ON person(birthday)`,
	}

	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			log.Fatalf("migration failed: %v\nSQL: %s", err, s)
		}
	}

	// if hasColumn(db, "person", "birthday_type") {
	// 	tx, err := db.Begin()
	// 	if err != nil {
	// 		log.Fatalf("migration failed: %v", err)
	// 	}
	// 	defer tx.Rollback()

	// 	if _, err := tx.Exec(`CREATE TABLE person_new (
	// 		id          INTEGER PRIMARY KEY AUTOINCREMENT,
	// 		family_id   INTEGER NOT NULL,
	// 		name        TEXT NOT NULL,
	// 		gender      TEXT CHECK(gender IN ('male','female','unknown')) DEFAULT 'unknown',
	// 		birthday    TEXT,
	// 		generation  INTEGER,
	// 		photo_url   TEXT,
	// 		phone       TEXT,
	// 		address     TEXT,
	// 		notes       TEXT,
	// 		is_alive    INTEGER DEFAULT 1,
	// 		created_at  DATETIME DEFAULT (datetime('now','localtime')),
	// 		updated_at  DATETIME DEFAULT (datetime('now','localtime')),
	// 		FOREIGN KEY (family_id) REFERENCES family(id) ON DELETE CASCADE
	// 	)`); err != nil {
	// 		log.Fatalf("migration failed: %v", err)
	// 	}

	// 	if _, err := tx.Exec(`INSERT INTO person_new (id, family_id, name, gender, birthday, generation, photo_url, phone, address, notes, is_alive, created_at, updated_at)
	// 		SELECT id, family_id, name, gender, birthday, generation, photo_url, phone, address, notes, is_alive, created_at, updated_at
	// 		FROM person`); err != nil {
	// 		log.Fatalf("migration failed: %v", err)
	// 	}

	// 	if _, err := tx.Exec(`DROP TABLE person`); err != nil {
	// 		log.Fatalf("migration failed: %v", err)
	// 	}
	// 	if _, err := tx.Exec(`ALTER TABLE person_new RENAME TO person`); err != nil {
	// 		log.Fatalf("migration failed: %v", err)
	// 	}
	// 	if _, err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_person_family ON person(family_id)`); err != nil {
	// 		log.Fatalf("migration failed: %v", err)
	// 	}
	// 	if _, err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_person_birthday ON person(birthday)`); err != nil {
	// 		log.Fatalf("migration failed: %v", err)
	// 	}

	// 	if err := tx.Commit(); err != nil {
	// 		log.Fatalf("migration failed: %v", err)
	// 	}
	// }

	log.Println("database migration complete")
}

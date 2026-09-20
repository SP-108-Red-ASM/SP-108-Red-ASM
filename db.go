package main

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

func open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	err = schema(db)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func schema(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS scan_cycle (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			started_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS host (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			cyc_id INTEGER NOT NULL,
			ip TEXT NOT NULL,
			state TEXT NOT NULL,
			FOREIGN KEY(cyc_id) REFERENCES scan_cycle(id)
		)`,
		`CREATE TABLE IF NOT EXISTS port_service (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			host_id INTEGER NOT NULL,
			port INTEGER NOT NULL,
			proto TEXT NOT NULL,
			service TEXT NOT NULL,
			state TEXT NOT NULL,
			FOREIGN KEY(host_id) REFERENCES host(id)
		)`,
		`CREATE TABLE IF NOT EXISTS web_endpoint (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			host_id INTEGER NOT NULL,
			url TEXT NOT NULL,
			FOREIGN KEY(host_id) REFERENCES host(id)
		)`,
		`CREATE TABLE IF NOT EXISTS finding (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ep_id INTEGER NOT NULL,
			typ TEXT NOT NULL,
			det TEXT,
			raw TEXT,
			FOREIGN KEY(ep_id) REFERENCES web_endpoint(id)
		)`,
		`CREATE TABLE IF NOT EXISTS chg (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			cyc_id INTEGER NOT NULL,
			ent TEXT NOT NULL,
			key TEXT NOT NULL,
			typ TEXT NOT NULL,
			det TEXT,
			FOREIGN KEY(cyc_id) REFERENCES scan_cycle(id)
		)`,
	}
	for _, s := range stmts {
		_, err := db.Exec(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func newCycle(db *sql.DB) (int64, error) {
	res, err := db.Exec(`INSERT INTO scan_cycle (started_at) VALUES (?)`, time.Now().UTC())
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

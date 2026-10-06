package diff

import (
	"SP_108_Red_ASM/internal/model"
	"database/sql"
	"fmt"
)

func Diff(db *sql.DB, cyc int64) ([]model.Change, error) {
	prev, err := prevCycle(db, cyc)
	if err != nil {
		return nil, err
	}
	if prev == 0 {
		return nil, nil
	}

	var chgs []model.Change

	rows, err := db.Query(`
		SELECT ip FROM host WHERE cyc_id = ? AND ip NOT IN (SELECT ip FROM host WHERE cyc_id = ?)
	`, cyc, prev)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var ip string
		err = rows.Scan(&ip)
		if err != nil {
			rows.Close()
			return nil, err
		}
		chgs = append(chgs, model.Change{Cyc: cyc, Ent: "host", Key: ip, Typ: "new"})
	}
	rows.Close()

	rows, err = db.Query(`
		SELECT ip FROM host WHERE cyc_id = ? AND ip NOT IN (SELECT ip FROM host WHERE cyc_id = ?)
	`, prev, cyc)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var ip string
		err = rows.Scan(&ip)
		if err != nil {
			rows.Close()
			return nil, err
		}
		chgs = append(chgs, model.Change{Cyc: cyc, Ent: "host", Key: ip, Typ: "gone"})
	}
	rows.Close()

	rows, err = db.Query(`
		SELECT h.ip, p.port, p.service FROM port_service p
		JOIN host h ON h.id = p.host_id
		WHERE h.cyc_id = ?
		AND (h.ip, p.port) NOT IN (
			SELECT h2.ip, p2.port FROM port_service p2 JOIN host h2 ON h2.id = p2.host_id WHERE h2.cyc_id = ?
		)
	`, cyc, prev)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var ip, svc string
		var port int
		err = rows.Scan(&ip, &port, &svc)
		if err != nil {
			rows.Close()
			return nil, err
		}
		chgs = append(chgs, model.Change{Cyc: cyc, Ent: "port", Key: fmt.Sprintf("%s:%d", ip, port), Typ: "new", Det: svc})
	}
	rows.Close()

	for _, ch := range chgs {
		err = saveChange(db, ch)
		if err != nil {
			return nil, err
		}
	}
	return chgs, nil
}

func prevCycle(db *sql.DB, cyc int64) (int64, error) {
	var prev int64
	err := db.QueryRow(`SELECT id FROM scan_cycle WHERE id < ? ORDER BY id DESC LIMIT 1`, cyc).Scan(&prev)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return prev, nil
}

func saveChange(db *sql.DB, ch model.Change) error {
	_, err := db.Exec(`INSERT INTO chg (cyc_id, ent, key, typ, det) VALUES (?,?,?,?,?)`, ch.Cyc, ch.Ent, ch.Key, ch.Typ, ch.Det)
	return err
}

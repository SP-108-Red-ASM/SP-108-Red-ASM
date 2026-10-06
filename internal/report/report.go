package report

import (
	"SP_108_Red_ASM/internal/model"
	"database/sql"
	"fmt"
)

func Report(db *sql.DB, cyc int64, chgs []model.Change) error {
	if len(chgs) == 0 {
		fmt.Printf("cycle %d: no changes detected\n", cyc)
	}
	for _, ch := range chgs {
		fmt.Printf("cycle %d [%s] %s %s: %s\n", cyc, ch.Typ, ch.Ent, ch.Key, ch.Det)
	}

	var n int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM finding f
		JOIN web_endpoint e ON e.id = f.ep_id
		JOIN host h ON h.id = e.host_id
		WHERE h.cyc_id = ?
	`, cyc).Scan(&n)
	if err != nil {
		return err
	}
	if n > 0 {
		fmt.Printf("ALERT: cycle %d has %d finding(s)\n", cyc, n)
	}
	return nil
}

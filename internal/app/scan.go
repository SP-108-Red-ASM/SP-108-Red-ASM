package app

import (
	"SP_108_Red_ASM/internal/config"
	"SP_108_Red_ASM/internal/discovery"
	"SP_108_Red_ASM/internal/model"
	"SP_108_Red_ASM/internal/store"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const banner = `==================================================
  SP-108 :: Attack Surface Management
==================================================`

func Run() error {
	fmt.Println(banner)

	c, err := config.Cfg()
	if err != nil {
		return err
	}
	fmt.Printf("config loaded: %d target(s) in scope, %d worker(s)\n\n", len(c.Scope), c.Workers)

	db, err := store.Open(c.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	cyc, err := store.NewCycle(db)
	if err != nil {
		return err
	}

	hosts, err := discovery.Scan(c)
	if err != nil {
		return err
	}

	err = store.Persist(db, cyc, hosts)
	if err != nil {
		return err
	}

	summarize(hosts, cyc, c.DBPath)

	if os.Getenv("SP108_DEBUG") != "" {
		b, err := json.MarshalIndent(hosts, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println("\n--- raw host data (SP108_DEBUG) ---")
		fmt.Println(string(b))
	}

	// sqlmap/audit, diff, and web display are added from here once ready
	return nil
}

func summarize(hosts []model.Host, cyc int64, dbPath string) {
	up := 0
	ports := 0
	for _, h := range hosts {
		if h.State == "up" {
			up++
		}
		ports += len(h.Ports)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println(" SCAN SUMMARY")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Hosts up:            %d/%d\n", up, len(hosts))
	fmt.Printf("Open ports found:    %d\n", ports)
	fmt.Printf("Vuln findings:       0 (assessment module pending)\n")
	fmt.Println()

	for _, h := range hosts {
		fmt.Printf("%s  [%s]\n", h.IP, h.State)
		for _, p := range h.Ports {
			svc := p.Service
			if svc == "" {
				svc = "unknown"
			}
			ver := p.Version
			if ver == "" {
				ver = "-"
			}
			fmt.Printf("  %-5d/%-4s %-7s %-12s %s\n", p.Port, p.Proto, p.State, svc, ver)
		}
		fmt.Println()
	}

	fmt.Printf("saved to database: %s (cycle #%d)\n", dbPath, cyc)
	fmt.Println(strings.Repeat("=", 50))
}

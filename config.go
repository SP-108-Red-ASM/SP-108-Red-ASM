package main

import (
	"encoding/json"
	"errors"
	"os"
)

func cfg() (Cfg, error) {
	p := os.Getenv("SP108_CONFIG")
	if p == "" {
		p = "config.json"
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return Cfg{}, err
	}
	var c Cfg
	err = json.Unmarshal(b, &c)
	if err != nil {
		return Cfg{}, err
	}
	if len(c.Scope) == 0 {
		return Cfg{}, errors.New("empty scope: fail-closed, refusing to run")
	}
	if c.Workers <= 0 {
		c.Workers = 4
	}
	if c.DBPath == "" {
		c.DBPath = "sp108.db"
	}
	if c.OutDir == "" {
		c.OutDir = "sqlmap_out"
	}
	return c, nil
}

// fail-closed: unmatched target is denied by default (loop falls through to false)
func allowed(t string, scope []string) bool {
	for _, s := range scope {
		if s == t {
			return true
		}
	}
	return false
}

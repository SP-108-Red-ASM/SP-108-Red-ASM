package config

import (
	"encoding/json"
	"errors"
	"os"
	"slices"

	"SP_108_Red_ASM/internal/model"
)

func Cfg() (model.Cfg, error) {
	p := os.Getenv("SP108_CONFIG")
	if p == "" {
		p = "config.json"
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return model.Cfg{}, err
	}
	var c model.Cfg
	err = json.Unmarshal(b, &c)
	if err != nil {
		return model.Cfg{}, err
	}
	if len(c.Scope) == 0 {
		return model.Cfg{}, errors.New("empty scope: fail-closed, refusing to run")
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
func Allowed(t string, scope []string) bool {
	return slices.Contains(scope, t)
}

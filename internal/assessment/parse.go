package assessment

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// Finding represents one injection technique detected for a parameter.
type Finding struct {
	URL       string
	Parameter string
	Location  string
	Technique string
	Payload   string
}

type injection struct {
	Parameter string `json:"parameter"`
	Place     string `json:"place"`
	Data      map[string]struct {
		Title   string `json:"title"`
		Payload string `json:"payload"`
	} `json:"data"`
}

// normalizeValue supports structured JSON and Python-literal strings.
func normalizeValue(raw json.RawMessage) ([]byte, error) {
	var literal string

	// An object or array is already structured JSON.
	if err := json.Unmarshal(raw, &literal); err != nil {
		return raw, nil
	}

	// Some versions return JSON encoded inside a JSON string.
	if json.Valid([]byte(literal)) {
		return []byte(literal), nil
	}

	if len(literal) > 1_000_000 {
		return nil, fmt.Errorf("SQLMap value exceeds size limit")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const script = `
import ast
import json
import sys

value = ast.literal_eval(sys.stdin.read())
json.dump(value, sys.stdout)
`

	cmd := exec.CommandContext(ctx, "python3", "-I", "-c", script)
	cmd.Stdin = strings.NewReader(literal)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("convert SQLMap value: %w", err)
	}

	return output, nil
}

func ParseFindings(targetURL string, results ScanResults) ([]Finding, error) {
	if !results.Success {
		return nil, fmt.Errorf("cannot parse unsuccessful SQLMap response")
	}

	if len(results.Errors) > 0 {
		return nil, fmt.Errorf(
			"SQLMap reported %d error(s); assessment needs review",
			len(results.Errors),
		)
	}

	var findings []Finding

	for _, item := range results.Data {
		// Type 1 is the injection data in our observed API response.
		if item.Type != 1 {
			continue
		}

		if item.Status != 1 {
			return nil, fmt.Errorf("injection result is incomplete")
		}

		data, err := normalizeValue(item.Value)
		if err != nil {
			return nil, err
		}

		var injections []injection
		if err := json.Unmarshal(data, &injections); err != nil {
			return nil, fmt.Errorf("decode injection findings: %w", err)
		}

		for _, entry := range injections {
			if entry.Parameter == "" || entry.Place == "" ||
				len(entry.Data) == 0 {
				return nil, fmt.Errorf("injection record is missing required fields")
			}

			// Sort techniques so the printed order is consistent.
			keys := make([]string, 0, len(entry.Data))
			for key := range entry.Data {
				keys = append(keys, key)
			}
			sort.Strings(keys)

			for _, key := range keys {
				detail := entry.Data[key]
				if detail.Title == "" || detail.Payload == "" {
					return nil, fmt.Errorf("injection technique is missing evidence")
				}

				findings = append(findings, Finding{
					URL:       targetURL,
					Parameter: entry.Parameter,
					Location:  entry.Place,
					Technique: detail.Title,
					Payload:   detail.Payload,
				})
			}
		}
	}

	return findings, nil
}

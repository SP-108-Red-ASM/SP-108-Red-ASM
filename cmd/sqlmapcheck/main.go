package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type taskResponse struct {
	Success bool   `json:"success"`
	TaskID  string `json:"taskid"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("http://127.0.0.1:8775/task/new")
	if err != nil {
		return fmt.Errorf("connect to SQLMap API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SQLMap API returned HTTP %d", resp.StatusCode)
	}

	var task taskResponse
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return fmt.Errorf("decode SQLMap response: %w", err)
	}

	if !task.Success || task.TaskID == "" {
		return fmt.Errorf("SQLMap API did not create a task")
	}

	fmt.Println("Connected to SQLMap API")
	fmt.Println("Created task:", task.TaskID)
	return nil
}

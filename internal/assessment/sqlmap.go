package assessment

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type taskResponse struct {
	Success bool   `json:"success"`
	TaskID  string `json:"taskid"`
}

func (c *Client) NewTask() (string, error) {
	resp, err := c.http.Get(c.baseURL + "/task/new")
	if err != nil {
		return "", fmt.Errorf("connect to SQLMap API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("SQLMap API returned HTTP %d", resp.StatusCode)
	}

	var task taskResponse
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return "", fmt.Errorf("decode SQLMap response: %w", err)
	}

	if !task.Success || task.TaskID == "" {
		return "", fmt.Errorf("SQLMap API did not create a task")
	}

	return task.TaskID, nil
}

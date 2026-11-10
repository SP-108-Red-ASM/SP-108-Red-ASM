package assessment

import (
	"bytes"
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

// StartLocalTest starts an assessment against one of our two test endpoints.
func (c *Client) StartLocalTest(taskID, targetURL string) error {
	switch targetURL {
	case "http://127.0.0.1:8000/vulnerable?id=1",
		"http://127.0.0.1:8000/safe?id=1":
		// Explicitly permitted local test targets.
	default:
		return fmt.Errorf("target is not an approved local test endpoint")
	}

	if taskID == "" {
		return fmt.Errorf("task ID is required")
	}

	options := map[string]any{
		"url":             targetURL,
		"batch":           true,
		"level":           1,
		"risk":            1,
		"technique":       "B",
		"testParameter":   "id",
		"ignoreRedirects": true,
	}

	body, err := json.Marshal(options)
	if err != nil {
		return fmt.Errorf("encode assessment options: %w", err)
	}

	resp, err := c.http.Post(
		c.baseURL+"/scan/"+taskID+"/start",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("start SQLMap assessment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SQLMap start returned HTTP %d", resp.StatusCode)
	}

	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode assessment start response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("SQLMap rejected assessment: %s", result.Message)
	}

	return nil
}

type ScanStatus struct {
	Success    bool   `json:"success"`
	Status     string `json:"status"`
	ReturnCode *int   `json:"returncode"`
}

func (c *Client) Status(taskID string) (ScanStatus, error) {
	var status ScanStatus

	if taskID == "" {
		return status, fmt.Errorf("task ID is required")
	}

	resp, err := c.http.Get(c.baseURL + "/scan/" + taskID + "/status")
	if err != nil {
		return status, fmt.Errorf("get SQLMap status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return status, fmt.Errorf("SQLMap status returned HTTP %d",
			resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return status, fmt.Errorf("decode SQLMap status: %w", err)
	}

	if !status.Success {
		return status, fmt.Errorf("SQLMap could not retrieve task status")
	}

	return status, nil
}

type ScanDataItem struct {
	Status int             `json:"status"`
	Type   int             `json:"type"`
	Value  json.RawMessage `json:"value"`
}

type ScanResults struct {
	Success bool              `json:"success"`
	Data    []ScanDataItem    `json:"data"`
	Errors  []json.RawMessage `json:"error"`
}

func (c *Client) Results(taskID string) (ScanResults, error) {
	var results ScanResults

	if taskID == "" {
		return results, fmt.Errorf("task ID is required")
	}

	resp, err := c.http.Get(c.baseURL + "/scan/" + taskID + "/data")
	if err != nil {
		return results, fmt.Errorf("get SQLMap results: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return results, fmt.Errorf(
			"SQLMap results returned HTTP %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return results, fmt.Errorf("decode SQLMap results: %w", err)
	}

	if !results.Success {
		return results, fmt.Errorf("SQLMap could not retrieve results")
	}

	return results, nil
}

func (c *Client) DeleteTask(taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task ID is required")
	}

	resp, err := c.http.Get(c.baseURL + "/task/" + taskID + "/delete")
	if err != nil {
		return fmt.Errorf("delete SQLMap task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"SQLMap task deletion returned HTTP %d",
			resp.StatusCode,
		)
	}

	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode task deletion response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("SQLMap could not delete task: %s", result.Message)
	}

	return nil
}

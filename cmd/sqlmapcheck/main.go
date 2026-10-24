package main

import (
	"fmt"
	"os"
	"time"

	"SP_108_Red_ASM/internal/assessment"
)

func main() {
	client := assessment.NewClient("http://127.0.0.1:8775")

	taskID, err := client.NewTask()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Connected to SQLMap API")
	fmt.Println("Created task:", taskID)

	targetURL := "http://127.0.0.1:8000/vulnerable?id=1"

	if err := client.StartLocalTest(taskID, targetURL); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Assessment started:", targetURL)
	fmt.Println("Task ID:", taskID)

	for {
		status, err := client.Status(taskID)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		fmt.Println("Assessment status:", status.Status)

		if status.Status == "terminated" {
			if status.ReturnCode == nil {
				fmt.Fprintln(os.Stderr, "Assessment ended without an exit code")
				os.Exit(1)
			}

			if *status.ReturnCode != 0 {
				fmt.Fprintf(os.Stderr,
					"Assessment failed with exit code %d\n",
					*status.ReturnCode)
				os.Exit(1)
			}

			fmt.Println("Assessment finished successfully")
			break
		}

		if status.Status != "running" {
			fmt.Fprintln(os.Stderr,
				"Unexpected assessment status:", status.Status)
			os.Exit(1)
		}

		time.Sleep(2 * time.Second)
	}
	results, err := client.Results(taskID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Received %d result records\n", len(results.Data))

	for _, item := range results.Data {
		fmt.Printf("Record type %d: %s\n", item.Type, item.Value)
	}

	if err := client.DeleteTask(taskID); err != nil {
		fmt.Fprintln(os.Stderr, "Task cleanup failed:", err)
		os.Exit(1)
	}

	fmt.Println("SQLMap task cleaned up")

	if len(results.Errors) > 0 {
		fmt.Fprintln(os.Stderr, "SQLMap reported errors:")
		for _, item := range results.Errors {
			fmt.Fprintln(os.Stderr, string(item))
		}
		os.Exit(1)
	}
}

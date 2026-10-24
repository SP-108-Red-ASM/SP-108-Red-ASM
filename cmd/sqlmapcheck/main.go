package main

import (
	"fmt"
	"os"

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
}

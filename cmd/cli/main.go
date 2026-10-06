package main

import (
	"fmt"
	"os"

	"SP_108_Red_ASM/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

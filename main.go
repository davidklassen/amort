package main

import (
	"log/slog"
	"os"

	"github.com/davidklassen/amort/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

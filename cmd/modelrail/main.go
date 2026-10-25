// Command modelrail starts the gateway.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"modelrail/internal/app"
	"modelrail/internal/constant"
)

// main parses the configuration path and runs the application until a signal
// requests shutdown.
func main() {
	configPath := flag.String(constant.FlagConfig, constant.DefaultConfigPath, "path to the YAML configuration file")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, *configPath); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", constant.AppName, err)
		os.Exit(1)
	}
}

package main

import (
	"context"
	"fmt"
	start "github.com/eric-schulze/we_sync_bricks/internal/init"
	"os"
)

func main() {
	ctx := context.Background()
	if err := start.Start(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

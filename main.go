package main

import (
    start "github.com/eric-schulze/we_sync_bricks/internal/init"
    "context"
    "fmt"
    "os"
)

func main() {
	ctx := context.Background()
	if err := start.Start(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
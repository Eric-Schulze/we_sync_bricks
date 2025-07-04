package main

import (
	"os"

	"github.com/eric-schulze/we_sync_bricks/internal/common/services/db"
)

func main() {
	// Pass command line arguments to RunMigrate
	db.RunMigrate(os.Args[1:])
}
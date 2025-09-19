package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/bricklink"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/db"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
	init_pkg "github.com/eric-schulze/we_sync_bricks/internal/init"
)

func main() {
	// Parse command line flags
	var (
		userID   = flag.Int64("user", 1, "User ID to use for BrickLink API calls (default: 1)")
		showInfo = flag.Bool("info", false, "Show current sync info and exit")
		verbose  = flag.Bool("v", false, "Verbose output")
	)
	flag.Parse()

	// Initialize logging system
	logFilePath := "logs/sync-colors.log"
	logLevel := logger.LogInfo
	if *verbose {
		logLevel = logger.LogDebug
	}

	err := logger.InitializeDefaultLogger(logLevel, logFilePath)
	if err != nil {
		log.Printf("Failed to initialize logger: %v", err)
		// Fall back to standard logging
		log.SetPrefix("sync-colors: ")
		log.SetFlags(0)
	}

	logger.Info("Starting BrickLink Colors Sync Tool")

	// Create context
	ctx := context.Background()

	// Load configuration
	config, err := init_pkg.LoadConfig()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize database connection
	dbService, err := db.NewDBService(&config.DBConfig, &ctx)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	logger.Debug("Connected to database successfully")

	// Initialize client manager
	clientManager := bricklink.NewClientManager(30*time.Minute, 100, dbService)
	defer clientManager.Close()

	// Create colors sync service
	colorsSyncService := bricklink.NewColorsSyncService(clientManager, dbService)

	// If showing info, display current state and exit
	if *showInfo {
		showSyncInfo(colorsSyncService)
		return
	}

	// Get user for API calls (in a real system, you'd want better user management)
	user := &models.User{
		ID:       *userID,
		Username: fmt.Sprintf("sync-user-%d", *userID),
		Email:    fmt.Sprintf("sync-user-%d@example.com", *userID),
	}

	logger.Info("Starting colors sync", "user_id", user.ID)

	// Perform the sync
	startTime := time.Now()
	err = colorsSyncService.SyncColors(user)
	if err != nil {
		logger.Error("Colors sync failed", "error", err, "duration", time.Since(startTime))
		os.Exit(1)
	}

	duration := time.Since(startTime)
	logger.Info("Colors sync completed successfully", "duration", duration)

	// Show final sync info
	showSyncInfo(colorsSyncService)
}

func showSyncInfo(service *bricklink.ColorsSyncService) {
	logger.Info("Retrieving sync information...")

	info, err := service.GetLastSyncInfo()
	if err != nil {
		logger.Error("Failed to get sync info", "error", err)
		return
	}

	fmt.Println("\n=== BrickLink Colors Sync Info ===")
	fmt.Printf("Total colors in database: %v\n", info["total_colors"])

	if lastUpdated := info["last_updated"]; lastUpdated != nil {
		if timestamp, ok := lastUpdated.(time.Time); ok {
			fmt.Printf("Last updated: %s\n", timestamp.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("Last updated: %v\n", lastUpdated)
		}
	} else {
		fmt.Println("Last updated: Never")
	}
	fmt.Println("==================================")
}
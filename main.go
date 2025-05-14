package main

import (
	"context"
	"fmt"
	"log"

	"github.com/eric-schulze/we_sync_bricks/bricklink"
	"github.com/eric-schulze/we_sync_bricks/config"
	"github.com/eric-schulze/we_sync_bricks/db"
	"github.com/eric-schulze/we_sync_bricks/inventory"
	"github.com/eric-schulze/we_sync_bricks/orders"
)

type App struct {
	Context context.Context
	DBService db.DBService
}

func main() {
	// Create a root context with cancellation
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := BootstrapApp(rootCtx)
	if err != nil {
		panic(err)
	}

	// Set properties of the predefined Logger, including
	// the log entry prefix and a flag to disable printing
	// the time, source file, and line number.
	log.SetPrefix("Inventory Manager: ")
	log.SetFlags(0)

	//messages := retrieveInventory()
	messages := retrieveOrders()

	// If no error was returned, print the returned map of
	// messages to the console.
	fmt.Println("Success")
	fmt.Println(messages)
}

func BootstrapApp(context context.Context) (*App, error) {
	var app = App{Context: context}

	appConfig, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	dbService, err := db.NewDBService(appConfig.DBConfig, &app.Context)
	if err != nil {
		return nil, err
	}

	app.DBService = *dbService

	return &app, nil
}

func retrieveInventory() string{
	var blClient = bricklink.NewBricklinkInventoryClient()

	var inventoryService = inventory.NewInventoryService(blClient)

	messages, err := inventoryService.GetInventory()
	if err != nil {
		log.Fatal(err)
	}

	return messages
}

func retrieveOrders() string{
	var blClient = bricklink.NewBricklinkOrdersClient()

	var ordersService = orders.NewOrdersService(blClient)

	messages, err := ordersService.GetOrders()
	if err != nil {
		log.Fatal(err)
	}

	return messages
}

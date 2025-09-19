package models

import (
	"context"
	"html/template"
)

type AppSettings struct {
	BaseURL string `mapstructure:"base_url"`
}

type AppConfig struct {
	DBConfig    DBConfig    `mapstructure:"db"`
	AppSettings AppSettings `mapstructure:"app"`
}

type App struct {
	Context               context.Context
	DBService             *DBService
	Config                AppConfig
	Templates             *template.Template
	AuthHandler           interface{} // Will be *auth.AuthHandler
	ProfileHandler        interface{} // Will be *profile.ProfileHandler
	PartialMinifigHandler interface{} // Will be *partial_minifigs.PartialMinifigHandler
	DashboardHandler      interface{} // Will be *dashboard.DashboardHandler
	OrdersHandler         interface{} // Will be *orders.OrdersHandler
	WebhookHandler        interface{} // Will be *webhooks.WebhookHandler
}

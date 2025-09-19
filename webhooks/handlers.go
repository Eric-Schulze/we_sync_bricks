package webhooks

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
	"github.com/eric-schulze/we_sync_bricks/notifications"
	"github.com/go-chi/chi/v5"
)

type WebhookHandler struct {
	app                 *models.App
	notificationService notifications.NotificationService
}

func NewWebhookHandler(app *models.App, notificationService notifications.NotificationService) *WebhookHandler {
	return &WebhookHandler{
		app:                 app,
		notificationService: notificationService,
	}
}

// RegisterRoutes registers webhook endpoints
func (h *WebhookHandler) RegisterRoutes(router chi.Router) {
	router.Route("/api/webhooks", func(r chi.Router) {
		// BrickLink webhook endpoint with user ID parameter
		r.Post("/bricklink/{userID}", h.HandleBricklinkWebhook)
	})
}

// HandleBricklinkWebhook processes incoming BrickLink push notifications
func (h *WebhookHandler) HandleBricklinkWebhook(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL parameter
	userIDStr := chi.URLParam(r, "userID")
	if userIDStr == "" {
		logger.Error("BrickLink webhook called without user ID")
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	// Convert user ID to integer
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		logger.Error("Invalid user ID in BrickLink webhook: %s", userIDStr)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Log the webhook reception
	logger.Info("Received BrickLink webhook for user ID: %d", userID)

	// According to BrickLink documentation, the POST body is typically empty
	// and should be treated as a "prompt to call Get-Notifications"
	
	// TODO: Implement Get Notifications API call for this user
	// TODO: Process the notifications and update orders
	
	// For now, just acknowledge receipt
	err = h.processBricklinkNotifications(userID)
	if err != nil {
		logger.Error("Failed to process BrickLink notifications for user %d: %v", userID, err)
		http.Error(w, "Failed to process notifications", http.StatusInternalServerError)
		return
	}

	// Return success to BrickLink
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// processBricklinkNotifications fetches and processes notifications for a specific user
func (h *WebhookHandler) processBricklinkNotifications(userID int) error {
	logger.Info("Processing BrickLink notifications for user", "user_id", userID)
	
	// Delegate to the notification service for BrickLink provider
	err := h.notificationService.ProcessWebhookForUser(userID, "bricklink")
	if err != nil {
		return fmt.Errorf("failed to process BrickLink notifications for user %d: %w", userID, err)
	}
	
	logger.Info("Successfully processed BrickLink notifications", "user_id", userID)
	return nil
}

// GenerateWebhookURL creates the webhook URL for a specific user
func GenerateWebhookURL(baseURL string, userID int) string {
	return fmt.Sprintf("%s/api/webhooks/bricklink/%d", baseURL, userID)
}
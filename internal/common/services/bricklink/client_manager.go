package bricklink

import (
	"fmt"
	"sync"
	"time"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
	"github.com/eric-schulze/we_sync_bricks/internal/services"
	"github.com/eric-schulze/we_sync_bricks/utils/oauth"
)

// CatalogClientCacheEntry holds a catalog client with its expiration time
type CatalogClientCacheEntry struct {
	Client    *BricklinkCatalogClient
	ExpiresAt time.Time
	LastUsed  time.Time
}

// OrdersClientCacheEntry holds an orders client with its expiration time
type OrdersClientCacheEntry struct {
	Client    *BricklinkOrdersClient
	ExpiresAt time.Time
	LastUsed  time.Time
}

// ClientManager manages a pool of OAuth clients for different users
type ClientManager struct {
	catalogCache  map[int64]*CatalogClientCacheEntry
	ordersCache   map[int64]*OrdersClientCacheEntry
	mutex         sync.RWMutex
	cacheTTL      time.Duration
	cleanupTicker *time.Ticker
	maxClients    int
	dbService     models.DBService
}

// NewClientManager creates a new client manager with specified cache TTL and max clients
func NewClientManager(cacheTTL time.Duration, maxClients int, dbService models.DBService) *ClientManager {
	manager := &ClientManager{
		catalogCache: make(map[int64]*CatalogClientCacheEntry),
		ordersCache:  make(map[int64]*OrdersClientCacheEntry),
		cacheTTL:     cacheTTL,
		maxClients:   maxClients,
		dbService:    dbService,
	}

	// Start cleanup routine to remove expired clients
	manager.cleanupTicker = time.NewTicker(cacheTTL / 2)
	go manager.cleanupRoutine()

	logger.Info("ClientManager initialized", "cache_ttl", cacheTTL, "max_clients", maxClients, "service", "ClientManager")
	return manager
}

// GetClient returns a cached catalog client for the user or creates a new one
func (cm *ClientManager) GetClient(user *models.User) (*BricklinkCatalogClient, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Check if we have a valid cached client
	if entry, exists := cm.catalogCache[user.ID]; exists {
		if time.Now().Before(entry.ExpiresAt) {
			entry.LastUsed = time.Now()
			logger.Debug("Using cached Bricklink catalog client", "user_id", user.ID, "expires_at", entry.ExpiresAt)
			return entry.Client, nil
		} else {
			// Client expired, remove it
			delete(cm.catalogCache, user.ID)
			logger.Debug("Removed expired Bricklink catalog client", "user_id", user.ID)
		}
	}

	// Check cache size limit (count both catalog and orders clients)
	totalClients := len(cm.catalogCache) + len(cm.ordersCache)
	if totalClients >= cm.maxClients {
		cm.evictOldestCatalogClient()
	}

	// Create new client for user
	client, err := cm.createCatalogClientForUser(user)
	if err != nil {
		logger.Error("Failed to create Bricklink catalog client for user", "user_id", user.ID, "error", err)
		return nil, err
	}

	// Cache the new client
	cm.catalogCache[user.ID] = &CatalogClientCacheEntry{
		Client:    client,
		ExpiresAt: time.Now().Add(cm.cacheTTL),
		LastUsed:  time.Now(),
	}

	logger.Info("Created and cached new Bricklink catalog client", "user_id", user.ID, "expires_at", time.Now().Add(cm.cacheTTL))
	return client, nil
}

// GetOrdersClient returns a cached orders client for the user or creates a new one
func (cm *ClientManager) GetOrdersClient(user *models.User) (*BricklinkOrdersClient, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Check if we have a valid cached client
	if entry, exists := cm.ordersCache[user.ID]; exists {
		if time.Now().Before(entry.ExpiresAt) {
			entry.LastUsed = time.Now()
			logger.Debug("Using cached Bricklink orders client", "user_id", user.ID, "expires_at", entry.ExpiresAt)
			return entry.Client, nil
		} else {
			// Client expired, remove it
			delete(cm.ordersCache, user.ID)
			logger.Debug("Removed expired Bricklink orders client", "user_id", user.ID)
		}
	}

	// Check cache size limit (count both catalog and orders clients)
	totalClients := len(cm.catalogCache) + len(cm.ordersCache)
	if totalClients >= cm.maxClients {
		cm.evictOldestOrdersClient()
	}

	// Create new client for user
	client, err := cm.createOrdersClientForUser(user)
	if err != nil {
		logger.Error("Failed to create Bricklink orders client for user", "user_id", user.ID, "error", err)
		return nil, err
	}

	// Cache the new client
	cm.ordersCache[user.ID] = &OrdersClientCacheEntry{
		Client:    client,
		ExpiresAt: time.Now().Add(cm.cacheTTL),
		LastUsed:  time.Now(),
	}

	logger.Info("Created and cached new Bricklink orders client", "user_id", user.ID, "expires_at", time.Now().Add(cm.cacheTTL))
	return client, nil
}

// GetBaseClient creates a base BLClient for the specific user
func (cm *ClientManager) GetBaseClient(user *models.User) (*BLClient, error) {
	// Get user's OAuth credentials from database
	userCredentials, err := cm.getUserOAuthCredentials(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth credentials for user %d: %w", user.ID, err)
	}

	// Create OAuth client with user's credentials
	oauthClient := oauth.NewOAuthClient(
		userCredentials.ConsumerKey,
		userCredentials.ConsumerSecret,
		userCredentials.Token,
		userCredentials.TokenSecret,
	)

	client := &BLClient{
		oAuthClient: oauthClient,
	}

	logger.Debug("Created base OAuth client for user", "user_id", user.ID)
	return client, nil
}

// createCatalogClientForUser creates a new OAuth catalog client for the specific user
func (cm *ClientManager) createCatalogClientForUser(user *models.User) (*BricklinkCatalogClient, error) {
	// Get base client
	baseClient, err := cm.GetBaseClient(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create base client for user %d: %w", user.ID, err)
	}

	client := &BricklinkCatalogClient{
		BLClient: *baseClient,
	}

	logger.Debug("Created OAuth catalog client for user", "user_id", user.ID)
	return client, nil
}

// createOrdersClientForUser creates a new OAuth orders client for the specific user
func (cm *ClientManager) createOrdersClientForUser(user *models.User) (*BricklinkOrdersClient, error) {
	// Get base client
	baseClient, err := cm.GetBaseClient(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create base client for user %d: %w", user.ID, err)
	}

	client := &BricklinkOrdersClient{
		BLClient: *baseClient,
	}

	logger.Debug("Created OAuth orders client for user", "user_id", user.ID)
	return client, nil
}

// UserOAuthCredentials holds the OAuth credentials for a user
type UserOAuthCredentials struct {
	ConsumerKey    string
	ConsumerSecret string
	Token          string
	TokenSecret    string
}

// getUserOAuthCredentials retrieves OAuth credentials for a user from the database
func (cm *ClientManager) getUserOAuthCredentials(userID int64) (*UserOAuthCredentials, error) {
	logger.Debug("Looking up OAuth credentials for user", "user_id", userID, "provider", "bricklink", "service", "ClientManager")

	// Use user service to get OAuth credentials
	credentials, err := services.GetServices().User.GetUserOAuthCredentials(userID, "bricklink")
	if err != nil {
		logger.Error("Failed to retrieve OAuth credentials from database", "user_id", userID, "provider", "bricklink", "service", "ClientManager", "error", err)

		// Fall back to mock credentials for development
		logger.Warn("Using fallback mock OAuth credentials", "user_id", userID, "service", "ClientManager")
		return &UserOAuthCredentials{
			ConsumerKey:    "4ED3302C6D1644CEA64E511455D1467B", // Development fallback
			ConsumerSecret: "9A91A62C32E040EBA2B16694C5120C6F",
			Token:          "BA8A79AD3C624A53AD3F67C2E5C21B1F",
			TokenSecret:    "733B0F8729C64879BA4094CB2936FE05",
		}, nil
	}

	// If no credentials found, use fallback
	if credentials == nil {
		logger.Info("No OAuth credentials found for user, using fallback mock credentials", "user_id", userID, "service", "ClientManager")
		return &UserOAuthCredentials{
			ConsumerKey:    "4ED3302C6D1644CEA64E511455D1467B", // Development fallback
			ConsumerSecret: "9A91A62C32E040EBA2B16694C5120C6F",
			Token:          "BA8A79AD3C624A53AD3F67C2E5C21B1F",
			TokenSecret:    "733B0F8729C64879BA4094CB2936FE05",
		}, nil
	}

	logger.Debug("Successfully retrieved OAuth credentials from database", "user_id", userID, "provider", "bricklink", "service", "ClientManager")
	return &UserOAuthCredentials{
		ConsumerKey:    credentials.ConsumerKey,
		ConsumerSecret: credentials.ConsumerSecret,
		Token:          credentials.Token,
		TokenSecret:    credentials.TokenSecret,
	}, nil
}

// evictOldestCatalogClient removes the least recently used catalog client from cache
func (cm *ClientManager) evictOldestCatalogClient() {
	var oldestUserID int64
	var oldestTime time.Time

	for userID, entry := range cm.catalogCache {
		if oldestTime.IsZero() || entry.LastUsed.Before(oldestTime) {
			oldestTime = entry.LastUsed
			oldestUserID = userID
		}
	}

	if oldestUserID != 0 {
		delete(cm.catalogCache, oldestUserID)
		logger.Debug("Evicted oldest catalog client from cache", "user_id", oldestUserID, "last_used", oldestTime)
	}
}

// evictOldestOrdersClient removes the least recently used orders client from cache
func (cm *ClientManager) evictOldestOrdersClient() {
	var oldestUserID int64
	var oldestTime time.Time

	for userID, entry := range cm.ordersCache {
		if oldestTime.IsZero() || entry.LastUsed.Before(oldestTime) {
			oldestTime = entry.LastUsed
			oldestUserID = userID
		}
	}

	if oldestUserID != 0 {
		delete(cm.ordersCache, oldestUserID)
		logger.Debug("Evicted oldest orders client from cache", "user_id", oldestUserID, "last_used", oldestTime)
	}
}

// cleanupRoutine periodically removes expired clients
func (cm *ClientManager) cleanupRoutine() {
	for range cm.cleanupTicker.C {
		cm.mutex.Lock()
		now := time.Now()
		
		// Clean up expired catalog clients
		catalogExpiredUsers := make([]int64, 0)
		for userID, entry := range cm.catalogCache {
			if now.After(entry.ExpiresAt) {
				catalogExpiredUsers = append(catalogExpiredUsers, userID)
			}
		}
		for _, userID := range catalogExpiredUsers {
			delete(cm.catalogCache, userID)
		}

		// Clean up expired orders clients
		ordersExpiredUsers := make([]int64, 0)
		for userID, entry := range cm.ordersCache {
			if now.After(entry.ExpiresAt) {
				ordersExpiredUsers = append(ordersExpiredUsers, userID)
			}
		}
		for _, userID := range ordersExpiredUsers {
			delete(cm.ordersCache, userID)
		}

		totalExpired := len(catalogExpiredUsers) + len(ordersExpiredUsers)
		if totalExpired > 0 {
			totalRemaining := len(cm.catalogCache) + len(cm.ordersCache)
			logger.Debug("Cleaned up expired clients", 
				"catalog_expired", len(catalogExpiredUsers),
				"orders_expired", len(ordersExpiredUsers),
				"total_remaining", totalRemaining)
		}

		cm.mutex.Unlock()
	}
}

// ClearCache removes all cached clients (useful for testing or manual cleanup)
func (cm *ClientManager) ClearCache() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	catalogCount := len(cm.catalogCache)
	ordersCount := len(cm.ordersCache)
	
	cm.catalogCache = make(map[int64]*CatalogClientCacheEntry)
	cm.ordersCache = make(map[int64]*OrdersClientCacheEntry)
	
	totalCleared := catalogCount + ordersCount
	logger.Info("Cleared client caches", 
		"catalog_cleared", catalogCount,
		"orders_cleared", ordersCount,
		"total_cleared", totalCleared)
}

// GetCacheStats returns statistics about the cache
func (cm *ClientManager) GetCacheStats() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	catalogCount := len(cm.catalogCache)
	ordersCount := len(cm.ordersCache)
	
	return map[string]interface{}{
		"catalog_clients": catalogCount,
		"orders_clients":  ordersCount,
		"total_clients":   catalogCount + ordersCount,
		"max_clients":     cm.maxClients,
		"cache_ttl":       cm.cacheTTL.String(),
	}
}

// Close shuts down the client manager and cleanup routines
func (cm *ClientManager) Close() {
	if cm.cleanupTicker != nil {
		cm.cleanupTicker.Stop()
	}
	cm.ClearCache()
	logger.Info("ClientManager closed")
}

package bricklink

import (
	"fmt"
	"sync"
	"time"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/db"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
	"github.com/eric-schulze/we_sync_bricks/utils/oauth"
)

// ClientCacheEntry holds a client with its expiration time
type ClientCacheEntry struct {
	Client    *BricklinkCatalogClient
	ExpiresAt time.Time
	LastUsed  time.Time
}

// ClientManager manages a pool of OAuth clients for different users
type ClientManager struct {
	cache        map[int64]*ClientCacheEntry
	mutex        sync.RWMutex
	cacheTTL     time.Duration
	cleanupTicker *time.Ticker
	maxClients   int
	dbService    models.DBService
}

// NewClientManager creates a new client manager with specified cache TTL and max clients
func NewClientManager(cacheTTL time.Duration, maxClients int, dbService models.DBService) *ClientManager {
	manager := &ClientManager{
		cache:        make(map[int64]*ClientCacheEntry),
		cacheTTL:     cacheTTL,
		maxClients:   maxClients,
		dbService:    dbService,
	}
	
	// Start cleanup routine to remove expired clients
	manager.cleanupTicker = time.NewTicker(cacheTTL / 2)
	go manager.cleanupRoutine()
	
	logger.Info("ClientManager initialized", "cache_ttl", cacheTTL, "max_clients", maxClients)
	return manager
}

// GetClient returns a cached client for the user or creates a new one
func (cm *ClientManager) GetClient(user *models.User) (*BricklinkCatalogClient, error) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	// Check if we have a valid cached client
	if entry, exists := cm.cache[user.ID]; exists {
		if time.Now().Before(entry.ExpiresAt) {
			entry.LastUsed = time.Now()
			logger.Debug("Using cached Bricklink client", "user_id", user.ID, "expires_at", entry.ExpiresAt)
			return entry.Client, nil
		} else {
			// Client expired, remove it
			delete(cm.cache, user.ID)
			logger.Debug("Removed expired Bricklink client", "user_id", user.ID)
		}
	}
	
	// Check cache size limit
	if len(cm.cache) >= cm.maxClients {
		cm.evictOldestClient()
	}
	
	// Create new client for user
	client, err := cm.createClientForUser(user)
	if err != nil {
		logger.Error("Failed to create Bricklink client for user", "user_id", user.ID, "error", err)
		return nil, err
	}
	
	// Cache the new client
	cm.cache[user.ID] = &ClientCacheEntry{
		Client:    client,
		ExpiresAt: time.Now().Add(cm.cacheTTL),
		LastUsed:  time.Now(),
	}
	
	logger.Info("Created and cached new Bricklink client", "user_id", user.ID, "expires_at", time.Now().Add(cm.cacheTTL))
	return client, nil
}

// createClientForUser creates a new OAuth client for the specific user
func (cm *ClientManager) createClientForUser(user *models.User) (*BricklinkCatalogClient, error) {
	// TODO: Get user's OAuth credentials from database
	// For now, we'll use placeholder logic
	
	// This is where you'd query the database for user-specific OAuth credentials
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
	
	client := &BricklinkCatalogClient{
		BLClient: BLClient{
			oAuthClient: oauthClient,
		},
	}
	
	logger.Debug("Created OAuth client for user", "user_id", user.ID)
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
	logger.Debug("Looking up OAuth credentials for user", "user_id", userID, "provider", "bricklink")
	
	sql := `
		SELECT consumer_key, consumer_secret, token, token_secret 
		FROM user_oauth_credentials 
		WHERE user_id = $1 AND provider = 'bricklink' AND is_active = true
	`
	
	// Query the database for user's OAuth credentials
	credentials, err := db.QueryRowToStructFromService[UserOAuthCredentials](cm.dbService, sql, userID)
	if err != nil {
		logger.Error("Failed to retrieve OAuth credentials from database", "user_id", userID, "provider", "bricklink", "error", err)
		
		// Fall back to mock credentials for development
		logger.Warn("Using fallback mock OAuth credentials", "user_id", userID)
		return &UserOAuthCredentials{
			ConsumerKey:    "4ED3302C6D1644CEA64E511455D1467B", // Development fallback
			ConsumerSecret: "9A91A62C32E040EBA2B16694C5120C6F",
			Token:          "BA8A79AD3C624A53AD3F67C2E5C21B1F", 
			TokenSecret:    "733B0F8729C64879BA4094CB2936FE05",
		}, nil
	}
	
	logger.Debug("Successfully retrieved OAuth credentials from database", "user_id", userID, "provider", "bricklink")
	return &credentials, nil
}

// evictOldestClient removes the least recently used client from cache
func (cm *ClientManager) evictOldestClient() {
	var oldestUserID int64
	var oldestTime time.Time
	
	for userID, entry := range cm.cache {
		if oldestTime.IsZero() || entry.LastUsed.Before(oldestTime) {
			oldestTime = entry.LastUsed
			oldestUserID = userID
		}
	}
	
	if oldestUserID != 0 {
		delete(cm.cache, oldestUserID)
		logger.Debug("Evicted oldest client from cache", "user_id", oldestUserID, "last_used", oldestTime)
	}
}

// cleanupRoutine periodically removes expired clients
func (cm *ClientManager) cleanupRoutine() {
	for range cm.cleanupTicker.C {
		cm.mutex.Lock()
		now := time.Now()
		expiredUsers := make([]int64, 0)
		
		for userID, entry := range cm.cache {
			if now.After(entry.ExpiresAt) {
				expiredUsers = append(expiredUsers, userID)
			}
		}
		
		for _, userID := range expiredUsers {
			delete(cm.cache, userID)
		}
		
		if len(expiredUsers) > 0 {
			logger.Debug("Cleaned up expired clients", "expired_count", len(expiredUsers), "remaining_count", len(cm.cache))
		}
		
		cm.mutex.Unlock()
	}
}

// ClearCache removes all cached clients (useful for testing or manual cleanup)
func (cm *ClientManager) ClearCache() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	count := len(cm.cache)
	cm.cache = make(map[int64]*ClientCacheEntry)
	logger.Info("Cleared client cache", "cleared_count", count)
}

// GetCacheStats returns statistics about the cache
func (cm *ClientManager) GetCacheStats() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	return map[string]interface{}{
		"cached_clients": len(cm.cache),
		"max_clients":    cm.maxClients,
		"cache_ttl":      cm.cacheTTL.String(),
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
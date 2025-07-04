package dashboard

import (
	"errors"
	"html/template"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/partial_minifigs"
)

type DashboardService struct {
	repo *DashboardRepository
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(repo *DashboardRepository) *DashboardService {
	return &DashboardService{
		repo: repo,
	}
}

// GetDashboardData retrieves all dashboard data for a user
func (s *DashboardService) GetDashboardData(userID int64) (*DashboardData, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	// Get user stats
	stats, err := s.repo.GetUserStats(userID)
	if err != nil {
		return nil, errors.New("failed to retrieve user stats")
	}

	// Get recent items
	recentItems, err := s.repo.GetRecentItems(userID, 5)
	if err != nil {
		return nil, errors.New("failed to retrieve recent items")
	}

	// Get user dashboard info
	userInfo, err := s.repo.GetUserDashboardInfo(userID)
	if err != nil {
		return nil, errors.New("failed to retrieve user dashboard info")
	}

	// Generate quick links
	quickLinks := s.generateQuickLinks()

	return &DashboardData{
		Stats:       *stats,
		RecentItems: recentItems,
		QuickLinks:  quickLinks,
		UserInfo:    *userInfo,
	}, nil
}

// GetUserStats retrieves just the statistics for a user
func (s *DashboardService) GetUserStats(userID int64) (*DashboardStats, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	stats, err := s.repo.GetUserStats(userID)
	if err != nil {
		return nil, errors.New("failed to retrieve user stats")
	}

	return stats, nil
}

// GetRecentActivity retrieves recent user activity
func (s *DashboardService) GetRecentActivity(userID int64, limit int) ([]ActivityItem, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	if limit <= 0 || limit > 50 {
		limit = 10 // Default limit
	}

	activity, err := s.repo.GetUserActivity(userID, limit)
	if err != nil {
		return nil, errors.New("failed to retrieve user activity")
	}

	return activity, nil
}

// GetRecentLists retrieves recent lists for a user
func (s *DashboardService) GetRecentLists(userID int64, limit int) ([]*partial_minifigs.PartialMinifigList, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	if limit <= 0 || limit > 50 {
		limit = 10 // Default limit
	}

	recentLists, err := s.repo.GetRecentLists(userID, limit)
	if err != nil {
		return nil, errors.New("failed to retrieve recent lists")
	}

	return recentLists, nil
}

// generateQuickLinks creates quick action links for the dashboard
func (s *DashboardService) generateQuickLinks() []QuickLink {
	return []QuickLink{
		{
			Title:       "Create New List",
			Description: "Start a new partial minifig collection",
			URL:         "/partial-minifigs-lists/new",
			Icon:        "plus-circle",
		},
		{
			Title:       "Browse Catalog",
			Description: "Search for LEGO minifigs and parts",
			URL:         "/catalog",
			Icon:        "search",
		},
		{
			Title:       "View Profile",
			Description: "Manage your account settings",
			URL:         "/profile",
			Icon:        "user",
		},
		{
			Title:       "Import Data",
			Description: "Import from BrickLink or other sources",
			URL:         "/import",
			Icon:        "upload",
		},
	}
}

// RefreshDashboardData refreshes cached dashboard data (if caching is implemented)
func (s *DashboardService) RefreshDashboardData(userID int64) error {
	// TODO: Implement cache invalidation if needed
	// For now, this is a no-op since we're not caching
	return nil
}

// InitializeDashboardHandler creates a fully initialized dashboard handler with all dependencies
func InitializeDashboardHandler(dbService models.DBService, templates *template.Template, jwtSecret []byte) *DashboardHandler {
	repo := NewDashboardRepository(dbService)
	service := NewDashboardService(repo)
	handler := NewDashboardHandler(service, templates, jwtSecret)
	return handler
}
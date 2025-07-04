package dashboard

import (
	"time"
)

// DashboardData represents the main dashboard data
type DashboardData struct {
	Stats       DashboardStats       `json:"stats"`
	RecentItems []RecentItem         `json:"recent_items"`
	QuickLinks  []QuickLink          `json:"quick_links"`
	UserInfo    UserDashboardInfo    `json:"user_info"`
}

// DashboardStats represents key statistics for the dashboard
type DashboardStats struct {
	TotalLists           int `json:"total_lists"`
	TotalPartialMinifigs int `json:"total_partial_minifigs"`
	CompletedMinifigs    int `json:"completed_minifigs"`
	TotalParts           int `json:"total_parts"`
	TotalCost            float64 `json:"total_cost"`
}

// RecentItem represents recently accessed or modified items
type RecentItem struct {
	ID          int       `json:"id"`
	Type        string    `json:"type"` // "list", "minifig", etc.
	Name        string    `json:"name"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
	URL         string    `json:"url"`
}

// QuickLink represents quick action links on the dashboard
type QuickLink struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Icon        string `json:"icon"`
}

// UserDashboardInfo represents user-specific dashboard information
type UserDashboardInfo struct {
	DisplayName    string     `json:"display_name"`
	LastLoginAt    *time.Time `json:"last_login_at"`
	MemberSince    time.Time  `json:"member_since"`
	NotificationCount int     `json:"notification_count"`
}

// ActivityItem represents recent user activity
type ActivityItem struct {
	ID          int       `json:"id"`
	UserID      int64     `json:"user_id"`
	Action      string    `json:"action"` // "created", "updated", "completed", etc.
	EntityType  string    `json:"entity_type"` // "list", "minifig", etc.
	EntityID    int       `json:"entity_id"`
	EntityName  string    `json:"entity_name"`
	CreatedAt   time.Time `json:"created_at"`
}
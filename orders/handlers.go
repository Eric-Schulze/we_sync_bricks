package orders

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/auth"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/bricklink"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
)

type OrdersHandler struct {
	service   *OrdersService
	templates *template.Template
	jwtSecret []byte
}

// OrderFilters represents the URL-based state for table functionality
type OrderFilters struct {
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
	Sort     string `form:"sort"`
	Order    string `form:"order"`
	Status   string `form:"status"`
	Search   string `form:"search"`
	DateFrom string `form:"date_from"`
	DateTo   string `form:"date_to"`
}

// GetDefaults returns OrderFilters with default values
func (f *OrderFilters) GetDefaults() *OrderFilters {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 25
	}
	if f.Sort == "" {
		f.Sort = "date_ordered"
	}
	if f.Order == "" {
		f.Order = "desc"
	}
	return f
}

// ValidateSort ensures the sort field is allowed to prevent SQL injection
func (f *OrderFilters) ValidateSort() {
	allowedSortFields := map[string]bool{
		"bricklink_order_id":   true,
		"date_ordered":         true,
		"buyer_name":           true,
		"status":               true,
		"total_price":          true,
		"currency_code":        true,
		"buyer_order_count":    true,
		"total_count":          true,
		"unique_count":         true,
		"date_status_changed":  true,
		"seller_name":          true,
		"store_name":           true,
		"buyer_email":          true,
		"payment_method":       true,
		"shipping_method":      true,
		"created_at":           true,
		"updated_at":           true,
	}
	
	if !allowedSortFields[f.Sort] {
		f.Sort = "date_ordered"
	}
	
	if f.Order != "asc" && f.Order != "desc" {
		f.Order = "desc"
	}
}

// PaginationData contains pagination information for templates
type PaginationData struct {
	Current     int
	Total       int
	PerPage     int
	TotalItems  int64
	HasPrev     bool
	HasNext     bool
	PrevPage    int
	NextPage    int
	FirstPage   int
	LastPage    int
	Pages       []int
	ShowingFrom int
	ShowingTo   int
}

// CalculatePagination creates pagination data from current state
func CalculatePagination(page, limit int, totalItems int64) PaginationData {
	totalPages := int((totalItems + int64(limit) - 1) / int64(limit))
	
	data := PaginationData{
		Current:    page,
		Total:      totalPages,
		PerPage:    limit,
		TotalItems: totalItems,
		HasPrev:    page > 1,
		HasNext:    page < totalPages,
		PrevPage:   page - 1,
		NextPage:   page + 1,
		FirstPage:  1,
		LastPage:   totalPages,
	}
	
	// Calculate showing range
	data.ShowingFrom = (page-1)*limit + 1
	data.ShowingTo = page * limit
	if data.ShowingTo > int(totalItems) {
		data.ShowingTo = int(totalItems)
	}
	
	// Calculate page numbers to show (show 5 pages around current)
	start := page - 2
	end := page + 2
	
	if start < 1 {
		start = 1
		end = 5
	}
	if end > totalPages {
		end = totalPages
		start = totalPages - 4
		if start < 1 {
			start = 1
		}
	}
	
	for i := start; i <= end; i++ {
		data.Pages = append(data.Pages, i)
	}
	
	return data
}

// SortData contains sorting information for templates
type SortData struct {
	Field string
	Order string
}

// NextOrder returns the opposite order for toggle functionality
func (s *SortData) NextOrder(field string) string {
	if s.Field == field && s.Order == "asc" {
		return "desc"
	}
	return "asc"
}

// SortClass returns just the sort state class
func (s *SortData) SortClass(field string) string {
	if s.Field != field {
		return ""
	}
	if s.Order == "asc" {
		return "sorted-asc"
	}
	return "sorted-desc"
}

// IsActive returns true if this field is currently being sorted
func (s *SortData) IsActive(field string) bool {
	return s.Field == field
}

// ParseOrderFilters extracts and validates URL parameters into OrderFilters
func ParseOrderFilters(r *http.Request) *OrderFilters {
	filters := &OrderFilters{}
	
	// Parse page
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		}
	}
	
	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			filters.Limit = limit
		}
	}
	
	// Parse sort and order
	filters.Sort = strings.TrimSpace(r.URL.Query().Get("sort"))
	filters.Order = strings.ToLower(strings.TrimSpace(r.URL.Query().Get("order")))
	
	// Parse other filters
	filters.Status = strings.TrimSpace(r.URL.Query().Get("status"))
	filters.Search = strings.TrimSpace(r.URL.Query().Get("search"))
	filters.DateFrom = strings.TrimSpace(r.URL.Query().Get("date_from"))
	filters.DateTo = strings.TrimSpace(r.URL.Query().Get("date_to"))
	
	// Apply defaults and validation
	filters.GetDefaults()
	filters.ValidateSort()
	
	return filters
}

// getMapKeys returns a slice of keys from a map[string]interface{} for logging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func NewOrdersHandler(service *OrdersService, templates *template.Template, jwtSecret []byte) *OrdersHandler {
	return &OrdersHandler{
		service:   service,
		templates: templates,
		jwtSecret: jwtSecret,
	}
}

// HandleOrdersPage renders the orders page with filtering, sorting, and pagination
func (handler *OrdersHandler) HandleOrdersPage() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Starting HandleOrdersPage request", "method", r.Method, "url", r.URL.String(), "remote_addr", r.RemoteAddr)

		// Get current user from context (set by auth middleware)
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			logger.Error("User not found in context - authentication failed")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		logger.Debug("User authenticated successfully", "user_id", user.ID, "email", user.Email)

		// Parse URL parameters for filtering/sorting/pagination
		filters := ParseOrderFilters(r)
		logger.Debug("Parsed URL parameters", 
			"user_id", user.ID, 
			"page", filters.Page, 
			"limit", filters.Limit,
			"sort", filters.Sort, 
			"order", filters.Order,
			"status", filters.Status,
			"search", filters.Search,
			"date_from", filters.DateFrom,
			"date_to", filters.DateTo)

		// Get filtered orders
		logger.Debug("Fetching filtered orders from service", "user_id", user.ID)
		orders, totalCount, err := handler.service.GetFilteredOrders(user.ID, *filters)
		
		if err == nil {
			logger.Debug("Successfully retrieved orders", 
				"user_id", user.ID, 
				"orders_count", len(orders), 
				"total_count", totalCount,
				"page", filters.Page,
				"limit", filters.Limit)
				
			// Log sample order data if we have orders (for debugging)
			if len(orders) > 0 {
				sampleOrder := orders[0]
				logger.Debug("Sample order data", 
					"user_id", user.ID,
					"order_id", sampleOrder.BricklinkOrderID,
					"buyer_name", sampleOrder.BuyerName,
					"total_count", sampleOrder.TotalCount,
					"unique_count", sampleOrder.UniqueCount,
					"buyer_order_count", sampleOrder.BuyerOrderCount,
					"status", sampleOrder.Status,
					"date_ordered", sampleOrder.DateOrdered)
			}
		}
		if err != nil {
			logger.Error("Failed to retrieve filtered orders", "user_id", user.ID, "error", err)
			
			data := map[string]interface{}{
				"Title":       "Orders",
				"CurrentPage": "orders",
				"Error":       "Failed to load orders",
				"Orders":      []interface{}{},
				"OrdersCount": 0,
				"User":        user,
				"Filters":     *filters,
				"Pagination":  CalculatePagination(1, 25, 0),
				"Sort":        &SortData{Field: filters.Sort, Order: filters.Order},
			}
			
			// Check if this is an HTMX request
			isHTMX := r.Header.Get("HX-Request") == "true"
			
			if isHTMX {
				// For HTMX requests, return only the orders-content template
				if err := handler.templates.ExecuteTemplate(w, "orders-content", data); err != nil {
					logger.Error("Template execution error for HTMX request", "template", "orders-content", "error", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			} else {
				// For full page requests, return the complete page
				if err := handler.templates.ExecuteTemplate(w, "orders", data); err != nil {
					logger.Error("Template execution error for full page request", "template", "orders", "error", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}
			return
		}

		// Calculate pagination data
		pagination := CalculatePagination(filters.Page, filters.Limit, totalCount)
		logger.Debug("Calculated pagination", 
			"user_id", user.ID,
			"current_page", pagination.Current,
			"total_pages", pagination.Total,
			"has_prev", pagination.HasPrev,
			"has_next", pagination.HasNext,
			"showing_from", pagination.ShowingFrom,
			"showing_to", pagination.ShowingTo)
		
		// Create sort data
		sortData := &SortData{Field: filters.Sort, Order: filters.Order}
		logger.Debug("Created sort data", 
			"user_id", user.ID,
			"sort_field", sortData.Field,
			"sort_order", sortData.Order,
			"sort_type", fmt.Sprintf("%T", sortData))

		data := map[string]interface{}{
			"Title":       "Orders",
			"CurrentPage": "orders",
			"Orders":      orders,
			"OrdersCount": totalCount,
			"User":        user,
			"Filters":     *filters,
			"Pagination":  pagination,
			"Sort":        sortData,
		}
		
		logger.Debug("Prepared template data", 
			"user_id", user.ID,
			"data_keys", getMapKeys(data),
			"orders_in_data", len(orders),
			"template_data_ready", true)

		// Check if this is an HTMX request
		isHTMX := r.Header.Get("HX-Request") == "true"
		logger.Debug("Request type determined", "user_id", user.ID, "is_htmx", isHTMX)

		if isHTMX {
			// For HTMX requests (like sorting), return only the orders_table template
			// since they target #orders-table-content specifically
			logger.Debug("Executing HTMX template", "user_id", user.ID, "template", "orders_table")
			
			// Debug the exact data being passed to template
			if sortVal, ok := data["Sort"]; ok {
				logger.Debug("Template data Sort field", 
					"user_id", user.ID,
					"sort_value", sortVal,
					"sort_type", fmt.Sprintf("%T", sortVal))
				if sortData, ok2 := sortVal.(SortData); ok2 {
					logger.Debug("Sort data methods available", 
						"user_id", user.ID,
						"field", sortData.Field,
						"order", sortData.Order,
						"class", sortData.SortClass("bricklink_order_id"))
				} else {
					logger.Error("Sort field is not SortData type", 
						"user_id", user.ID,
						"actual_type", fmt.Sprintf("%T", sortVal))
				}
			} else {
				logger.Error("Sort field missing from template data", "user_id", user.ID)
			}
			
			if err := handler.templates.ExecuteTemplate(w, "orders_table", data); err != nil {
				logger.Error("Template execution error for HTMX request", 
					"user_id", user.ID, 
					"template", "orders_table", 
					"error", err,
					"error_string", err.Error())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// For full page requests, return the complete page
		logger.Debug("Executing full page template", "user_id", user.ID, "template", "orders")
		
		// Debug the exact data being passed to template for full page too
		if sortVal, ok := data["Sort"]; ok {
			logger.Debug("Full page template data Sort field", 
				"user_id", user.ID,
				"sort_value", sortVal,
				"sort_type", fmt.Sprintf("%T", sortVal))
		} else {
			logger.Error("Sort field missing from full page template data", "user_id", user.ID)
		}
		
		if err := handler.templates.ExecuteTemplate(w, "orders", data); err != nil {
			logger.Error("Template execution error for full page request", "user_id", user.ID, "template", "orders", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		logger.Debug("Successfully rendered orders page", "user_id", user.ID, "orders_count", totalCount, "page", filters.Page)
	})
}

// HandleRefreshOrders handles the refresh orders endpoint
func (handler *OrdersHandler) HandleRefreshOrders() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Starting HandleRefreshOrders request", "method", r.Method, "url", r.URL.String(), "remote_addr", r.RemoteAddr)

		// Get current user from context (set by auth middleware)
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			logger.Error("User not found in context - authentication failed")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		logger.Info("Starting order sync request", "user_id", user.ID)

		// Start the sync process
		orderSync, err := handler.service.SyncOrdersFromBrickLink(user)
		if err != nil {
			logger.Error("Failed to sync orders from BrickLink", "user_id", user.ID, "error", err)
			
			// Get current orders and count for error display
			orders, _ := handler.service.GetOrders(user.ID)
			if orders == nil {
				orders = []bricklink.Order{}
			}
			ordersCount, _ := handler.service.GetOrdersCount(user.ID)
			
			// Return failed sync status
			w.Header().Set("Content-Type", "text/html")
			
			// Create default filters and sorting for template consistency
			defaultFilters := &OrderFilters{}
			defaultFilters.GetDefaults()
			sortData := SortData{Field: defaultFilters.Sort, Order: defaultFilters.Order}
			pagination := CalculatePagination(1, 25, int64(len(orders)))
			
			data := map[string]interface{}{
				"Orders":      orders,
				"OrdersCount": ordersCount,
				"SyncStatus":  "failed",
				"SyncError":   err.Error(),
				"SyncCount":   0,
				"Sort":        sortData,
				"Filters":     *defaultFilters,
				"Pagination":  pagination,
			}
			
			if err := handler.templates.ExecuteTemplate(w, "orders_table", data); err != nil {
				logger.Error("Failed to execute template for sync error", "template", "orders_table", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Get updated orders after sync
		orders, err := handler.service.GetOrders(user.ID)
		if err != nil {
			logger.Error("Failed to retrieve orders after sync", "user_id", user.ID, "error", err)
			orders = []bricklink.Order{} // Fallback to empty array
		}

		// Get updated count
		ordersCount, err := handler.service.GetOrdersCount(user.ID)
		if err != nil {
			logger.Error("Failed to get orders count after sync", "user_id", user.ID, "error", err)
			ordersCount = 0
		}

		// Set content type for HTMX response
		w.Header().Set("Content-Type", "text/html")

		// Return the updated orders table as HTML
		// Create default filters and sorting for template consistency
		defaultFilters := &OrderFilters{}
		defaultFilters.GetDefaults()
		sortData := SortData{Field: defaultFilters.Sort, Order: defaultFilters.Order}
		pagination := CalculatePagination(1, 25, ordersCount)
		
		data := map[string]interface{}{
			"Orders":      orders,
			"OrdersCount": ordersCount,
			"SyncStatus":  orderSync.SyncStatus,
			"SyncCount":   orderSync.OrdersCount,
			"SyncError":   nil,
			"Sort":        sortData,
			"Filters":     *defaultFilters,
			"Pagination":  pagination,
		}
		
		// Include error message if present
		if orderSync.ErrorMessage != nil {
			data["SyncError"] = *orderSync.ErrorMessage
		}

		if err := handler.templates.ExecuteTemplate(w, "orders_table", data); err != nil {
			logger.Error("Failed to execute template", "template", "orders_table", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		logger.Info("Successfully refreshed orders", "user_id", user.ID, "sync_status", orderSync.SyncStatus, "sync_count", orderSync.OrdersCount)
	})
}

// HandleOrdersAPI returns orders data as JSON for API requests
func (handler *OrdersHandler) HandleOrdersAPI() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Starting HandleOrdersAPI request", "method", r.Method, "url", r.URL.String(), "remote_addr", r.RemoteAddr)

		// Get current user from context (set by auth middleware)
		user, ok := r.Context().Value(auth.UserContextKey).(*models.User)
		if !ok {
			logger.Error("User not found in context - authentication failed")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse query parameters
		limitStr := r.URL.Query().Get("limit")
		if limitStr == "" {
			limitStr = "50"
		}
		offsetStr := r.URL.Query().Get("offset")
		if offsetStr == "" {
			offsetStr = "0"
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 50
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			offset = 0
		}

		// Get orders for the user
		orders, err := handler.service.GetOrders(user.ID)
		if err != nil {
			logger.Error("Failed to retrieve orders", "user_id", user.ID, "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Failed to retrieve orders",
			})
			return
		}

		// Apply pagination
		totalCount := len(orders)
		if offset >= totalCount {
			orders = []bricklink.Order{}
		} else {
			end := offset + limit
			if end > totalCount {
				end = totalCount
			}
			orders = orders[offset:end]
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"orders":     orders,
			"totalCount": totalCount,
			"limit":      limit,
			"offset":     offset,
			"hasMore":    offset+limit < totalCount,
		})

		logger.Debug("Successfully returned orders API response", "user_id", user.ID, "total_count", totalCount, "returned_count", len(orders))
	})
}
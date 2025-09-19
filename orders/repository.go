package orders

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eric-schulze/we_sync_bricks/internal/common/models"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/bricklink"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/db"
	"github.com/eric-schulze/we_sync_bricks/internal/common/services/logger"
)

// Helper functions for parsing BrickLink API data

func parseOrderDate(dateStr string) (time.Time, error) {
	// Try ISO format first
	if t, err := time.Parse("2006-01-02T15:04:05.999Z", dateStr); err == nil {
		return t, nil
	}
	// Try alternative format
	return time.Parse("2006-01-02", dateStr)
}

func parseFloatPointer(s *string) *float64 {
	if s == nil || *s == "" {
		return nil
	}
	if val, err := strconv.ParseFloat(*s, 64); err == nil {
		return &val
	}
	return nil
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

type OrderRepository struct {
	db models.DBService
}

func NewOrderRepository(dbService models.DBService) *OrderRepository {
	return &OrderRepository{
		db: dbService,
	}
}

// GetAllOrders retrieves all orders for a specific user from the database
func (repo *OrderRepository) GetAllOrders(userID int64) ([]bricklink.Order, error) {
	sql := `SELECT 
		id, user_id, bricklink_order_id, date_ordered, date_status_changed,
		seller_name, store_name, buyer_name, buyer_email, buyer_order_count,
		status, is_invoiced, require_insurance, remarks, total_count, unique_count,
		total_weight, is_filed, drive_thru_sent, payment_method, payment_currency_code,
		payment_status, date_paid, shipping_method_id, shipping_method, tracking_link,
		shipping_address_name, shipping_address_full, shipping_country_code,
		currency_code, subtotal, total_price, etc1, etc2, insurance_cost,
		shipping_cost, credit_amount, coupon_amount, vat_rate, vat_amount,
		display_currency_code, display_subtotal, display_grand_total,
		display_etc1, display_etc2, display_insurance, display_shipping,
		display_credit, display_coupon, display_vat_rate, display_vat_amount,
		created_at, updated_at
	FROM orders 
	WHERE user_id = $1 
	ORDER BY date_ordered DESC;`

	orders, err := db.CollectRowsToStructFromService[bricklink.Order](repo.db, sql, userID)
	if err != nil {
		logger.Error("Database Error while retrieving orders", "user_id", userID, "error", err)
		return nil, err
	}

	return orders, nil
}

// GetOrderByBricklinkID retrieves a specific order by its BrickLink order ID for a specific user
func (repo *OrderRepository) GetOrderByBricklinkID(bricklinkOrderID int, userID int64) (*bricklink.Order, error) {
	sql := `SELECT 
		id, user_id, bricklink_order_id, date_ordered, date_status_changed,
		seller_name, store_name, buyer_name, buyer_email, buyer_order_count,
		status, is_invoiced, require_insurance, remarks, total_count, unique_count,
		total_weight, is_filed, drive_thru_sent, payment_method, payment_currency_code,
		payment_status, date_paid, shipping_method_id, shipping_method, tracking_link,
		shipping_address_name, shipping_address_full, shipping_country_code,
		currency_code, subtotal, total_price, etc1, etc2, insurance_cost,
		shipping_cost, credit_amount, coupon_amount, vat_rate, vat_amount,
		display_currency_code, display_subtotal, display_grand_total,
		display_etc1, display_etc2, display_insurance, display_shipping,
		display_credit, display_coupon, display_vat_rate, display_vat_amount,
		created_at, updated_at
	FROM orders 
	WHERE bricklink_order_id = $1 AND user_id = $2;`

	order, err := db.QueryRowToStructFromService[bricklink.Order](repo.db, sql, bricklinkOrderID, userID)
	if err != nil {
		logger.Error("Database Error while retrieving order", "bricklink_order_id", bricklinkOrderID, "user_id", userID, "error", err)
		return nil, err
	}

	return &order, nil
}

// GetFilteredOrders retrieves orders with filtering, sorting, and pagination
func (repo *OrderRepository) GetFilteredOrders(userID int64, filters OrderFilters) ([]bricklink.Order, int64, error) {
	logger.Debug("Building filtered query", 
		"user_id", userID,
		"page", filters.Page,
		"limit", filters.Limit,
		"sort", filters.Sort,
		"order", filters.Order,
		"status", filters.Status,
		"search", filters.Search)
		
	// Build the WHERE clause and arguments
	whereClause := "WHERE user_id = $1"
	args := []interface{}{userID}
	argCount := 1
	
	// Add status filter
	if filters.Status != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filters.Status)
	}
	
	// Add search filter (search in buyer_name, store_name, seller_name, and order_id)
	if filters.Search != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND (buyer_name ILIKE $%d OR store_name ILIKE $%d OR seller_name ILIKE $%d OR CAST(bricklink_order_id AS TEXT) ILIKE $%d)", argCount, argCount, argCount, argCount)
		args = append(args, "%"+filters.Search+"%")
	}
	
	// Add date range filters
	if filters.DateFrom != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND date_ordered >= $%d", argCount)
		args = append(args, filters.DateFrom)
	}
	
	if filters.DateTo != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND date_ordered <= $%d", argCount)
		args = append(args, filters.DateTo)
	}
	
	// First, get the total count with improved query
	countSQL := "SELECT COUNT(*) FROM orders " + whereClause
	logger.Debug("Executing count query", "user_id", userID, "sql", countSQL, "args_count", len(args))
	
	rows, err := repo.db.CollectRowsToMap(countSQL, args...)
	if err != nil {
		logger.Error("Database Error while counting filtered orders", "user_id", userID, "sql", countSQL, "error", err)
		return nil, 0, err
	}
	
	var totalCount int64
	if len(rows) > 0 {
		if count, ok := rows[0]["count"]; ok {
			if countVal, ok := count.(int64); ok {
				totalCount = countVal
			}
		}
	}
	
	// Build the main query with all columns and improved ORDER BY
	sql := `SELECT 
		id, user_id, bricklink_order_id, date_ordered, date_status_changed,
		seller_name, store_name, buyer_name, buyer_email, buyer_order_count,
		status, is_invoiced, require_insurance, remarks, total_count, unique_count,
		total_weight, is_filed, drive_thru_sent, payment_method, payment_currency_code,
		payment_status, date_paid, shipping_method_id, shipping_method, tracking_link,
		shipping_address_name, shipping_address_full, shipping_country_code,
		currency_code, subtotal, total_price, etc1, etc2, insurance_cost,
		shipping_cost, credit_amount, coupon_amount, vat_rate, vat_amount,
		display_currency_code, display_subtotal, display_grand_total,
		display_etc1, display_etc2, display_insurance, display_shipping,
		display_credit, display_coupon, display_vat_rate, display_vat_amount,
		created_at, updated_at
	FROM orders ` + whereClause
	
	// Add ORDER BY clause with NULLS LAST for better performance
	orderBy := fmt.Sprintf(" ORDER BY %s %s NULLS LAST", filters.Sort, strings.ToUpper(filters.Order))
	sql += orderBy
	
	// Add LIMIT and OFFSET for pagination
	offset := (filters.Page - 1) * filters.Limit
	sql += fmt.Sprintf(" LIMIT %d OFFSET %d", filters.Limit, offset)
	
	logger.Debug("Executing main query", 
		"user_id", userID, 
		"final_sql_length", len(sql),
		"args_count", len(args),
		"limit", filters.Limit,
		"offset", offset)
	
	// Execute the query using the existing working method
	orders, err := db.CollectRowsToStructFromService[bricklink.Order](repo.db, sql, args...)
	if err != nil {
		logger.Error("Database Error while retrieving filtered orders", 
			"user_id", userID, 
			"sql_preview", sql[:min(200, len(sql))],
			"error", err)
		return nil, 0, err
	}
	
	logger.Debug("Query completed successfully", 
		"user_id", userID,
		"returned_orders", len(orders),
		"total_count", totalCount)
	
	return orders, totalCount, nil
}

// OrderActionResult represents the result of creating or updating an order
type OrderActionResult struct {
	Order   *bricklink.Order
	Created bool // true if created, false if updated
}

// CreateOrUpdateOrder creates a new order or updates an existing one
func (repo *OrderRepository) CreateOrUpdateOrder(userID int64, blOrder bricklink.BLOrder) (*OrderActionResult, error) {
	// Parse the date ordered
	dateOrdered, err := parseOrderDate(blOrder.DateOrdered)
	if err != nil {
		logger.Error("Failed to parse date_ordered", "date_ordered", blOrder.DateOrdered, "error", err)
		return nil, err
	}

	// Parse date status changed
	var dateStatusChanged *time.Time
	if blOrder.DateStatusChanged != nil && *blOrder.DateStatusChanged != "" {
		if parsed, err := parseOrderDate(*blOrder.DateStatusChanged); err == nil {
			dateStatusChanged = &parsed
		}
	}

	// Parse date paid
	var datePaid *time.Time
	if blOrder.Payment != nil && blOrder.Payment.DatePaid != nil && *blOrder.Payment.DatePaid != "" {
		if parsed, err := parseOrderDate(*blOrder.Payment.DatePaid); err == nil {
			datePaid = &parsed
		}
	}

	// Parse all order fields
	var totalPrice float64
	var currencyCode string
	
	// Parse cost information
	var subtotal, shippingCost, insuranceCost, creditAmount, couponAmount, etc1, etc2, vatRate, vatAmount *float64
	if blOrder.Cost != nil {
		currencyCode = getStringValue(blOrder.Cost.CurrencyCode)
		if blOrder.Cost.GrandTotal != nil {
			if parsed, err := strconv.ParseFloat(*blOrder.Cost.GrandTotal, 64); err == nil {
				totalPrice = parsed
			}
		}
		subtotal = parseFloatPointer(blOrder.Cost.Subtotal)
		shippingCost = parseFloatPointer(blOrder.Cost.Shipping)
		insuranceCost = parseFloatPointer(blOrder.Cost.Insurance)
		creditAmount = parseFloatPointer(blOrder.Cost.Credit)
		couponAmount = parseFloatPointer(blOrder.Cost.Coupon)
		etc1 = parseFloatPointer(blOrder.Cost.Etc1)
		etc2 = parseFloatPointer(blOrder.Cost.Etc2)
		vatRate = parseFloatPointer(blOrder.Cost.VatRate)
		vatAmount = parseFloatPointer(blOrder.Cost.VatAmount)
	}

	// Parse display cost information
	var displayCurrencyCode *string
	var displaySubtotal, displayGrandTotal, displayShipping, displayInsurance, displayCredit, displayCoupon, displayEtc1, displayEtc2, displayVatRate, displayVatAmount *float64
	if blOrder.DispCost != nil {
		displayCurrencyCode = blOrder.DispCost.CurrencyCode
		displaySubtotal = parseFloatPointer(blOrder.DispCost.Subtotal)
		displayGrandTotal = parseFloatPointer(blOrder.DispCost.GrandTotal)
		displayShipping = parseFloatPointer(blOrder.DispCost.Shipping)
		displayInsurance = parseFloatPointer(blOrder.DispCost.Insurance)
		displayCredit = parseFloatPointer(blOrder.DispCost.Credit)
		displayCoupon = parseFloatPointer(blOrder.DispCost.Coupon)
		displayEtc1 = parseFloatPointer(blOrder.DispCost.Etc1)
		displayEtc2 = parseFloatPointer(blOrder.DispCost.Etc2)
		displayVatRate = parseFloatPointer(blOrder.DispCost.VatRate)
		displayVatAmount = parseFloatPointer(blOrder.DispCost.VatAmount)
	}

	// Parse shipping information
	var shippingMethodID *int
	var shippingMethod, trackingLink, shippingAddressName, shippingAddressFull, shippingCountryCode *string
	if blOrder.Shipping != nil {
		shippingMethodID = blOrder.Shipping.MethodID
		shippingMethod = blOrder.Shipping.Method
		trackingLink = blOrder.Shipping.TrackingLink
		if blOrder.Shipping.Address != nil {
			shippingAddressFull = blOrder.Shipping.Address.Full
			shippingCountryCode = blOrder.Shipping.Address.CountryCode
			if blOrder.Shipping.Address.Name != nil && blOrder.Shipping.Address.Name.Full != nil {
				shippingAddressName = blOrder.Shipping.Address.Name.Full
			}
		}
	}

	// Parse payment information
	var paymentMethod, paymentCurrencyCode, paymentStatus *string
	if blOrder.Payment != nil {
		paymentMethod = blOrder.Payment.Method
		paymentCurrencyCode = blOrder.Payment.CurrencyCode
		paymentStatus = blOrder.Payment.Status
	}

	// First, check if the order already exists
	existingOrder, err := repo.GetOrderByBricklinkID(blOrder.OrderID, userID)
	if err == nil && existingOrder != nil {
		// Update existing order with all fields
		sql := `UPDATE orders SET 
				date_ordered = $1, date_status_changed = $2, seller_name = $3, store_name = $4,
				buyer_name = $5, buyer_email = $6, buyer_order_count = $7, status = $8,
				is_invoiced = $9, require_insurance = $10, remarks = $11, total_count = $12,
				unique_count = $13, total_weight = $14, is_filed = $15, drive_thru_sent = $16,
				payment_method = $17, payment_currency_code = $18, payment_status = $19, date_paid = $20,
				shipping_method_id = $21, shipping_method = $22, tracking_link = $23,
				shipping_address_name = $24, shipping_address_full = $25, shipping_country_code = $26,
				currency_code = $27, subtotal = $28, total_price = $29, etc1 = $30, etc2 = $31,
				insurance_cost = $32, shipping_cost = $33, credit_amount = $34, coupon_amount = $35,
				vat_rate = $36, vat_amount = $37, display_currency_code = $38, display_subtotal = $39,
				display_grand_total = $40, display_etc1 = $41, display_etc2 = $42,
				display_insurance = $43, display_shipping = $44, display_credit = $45,
				display_coupon = $46, display_vat_rate = $47, display_vat_amount = $48,
				updated_at = CURRENT_TIMESTAMP
			WHERE bricklink_order_id = $49 AND user_id = $50
			RETURNING id, user_id, bricklink_order_id, date_ordered, date_status_changed,
				seller_name, store_name, buyer_name, buyer_email, buyer_order_count,
				status, is_invoiced, require_insurance, remarks, total_count, unique_count,
				total_weight, is_filed, drive_thru_sent, payment_method, payment_currency_code,
				payment_status, date_paid, shipping_method_id, shipping_method, tracking_link,
				shipping_address_name, shipping_address_full, shipping_country_code,
				currency_code, subtotal, total_price, etc1, etc2, insurance_cost,
				shipping_cost, credit_amount, coupon_amount, vat_rate, vat_amount,
				display_currency_code, display_subtotal, display_grand_total,
				display_etc1, display_etc2, display_insurance, display_shipping,
				display_credit, display_coupon, display_vat_rate, display_vat_amount,
				created_at, updated_at;`

		updatedOrder, err := db.QueryRowToStructFromService[bricklink.Order](repo.db, sql,
			dateOrdered, dateStatusChanged, blOrder.SellerName, blOrder.StoreName,
			blOrder.BuyerName, blOrder.BuyerEmail, blOrder.BuyerOrderCount, blOrder.Status,
			blOrder.IsInvoiced, blOrder.RequireInsurance, blOrder.Remarks, blOrder.TotalCount,
			blOrder.UniqueCount, blOrder.TotalWeight, blOrder.IsFiled, blOrder.DriveThruSent,
			paymentMethod, paymentCurrencyCode, paymentStatus, datePaid,
			shippingMethodID, shippingMethod, trackingLink,
			shippingAddressName, shippingAddressFull, shippingCountryCode,
			currencyCode, subtotal, totalPrice, etc1, etc2,
			insuranceCost, shippingCost, creditAmount, couponAmount,
			vatRate, vatAmount, displayCurrencyCode, displaySubtotal,
			displayGrandTotal, displayEtc1, displayEtc2,
			displayInsurance, displayShipping, displayCredit,
			displayCoupon, displayVatRate, displayVatAmount,
			blOrder.OrderID, userID)
		if err != nil {
			logger.Error("Database Error while updating order", "bricklink_order_id", blOrder.OrderID, "user_id", userID, "error", err)
			return nil, err
		}

		logger.Info("Updated existing order", "bricklink_order_id", blOrder.OrderID, "user_id", userID)
		return &OrderActionResult{Order: &updatedOrder, Created: false}, nil
	}

	// Create new order with all fields
	sql := `INSERT INTO orders (
			user_id, bricklink_order_id, date_ordered, date_status_changed, seller_name, store_name,
			buyer_name, buyer_email, buyer_order_count, status, is_invoiced, require_insurance,
			remarks, total_count, unique_count, total_weight, is_filed, drive_thru_sent,
			payment_method, payment_currency_code, payment_status, date_paid,
			shipping_method_id, shipping_method, tracking_link, shipping_address_name,
			shipping_address_full, shipping_country_code, currency_code, subtotal,
			total_price, etc1, etc2, insurance_cost, shipping_cost, credit_amount,
			coupon_amount, vat_rate, vat_amount, display_currency_code, display_subtotal,
			display_grand_total, display_etc1, display_etc2, display_insurance,
			display_shipping, display_credit, display_coupon, display_vat_rate,
			display_vat_amount, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18,
			$19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34,
			$35, $36, $37, $38, $39, $40, $41, $42, $43, $44, $45, $46, $47, $48, $49, $50, CURRENT_TIMESTAMP
		) RETURNING id, user_id, bricklink_order_id, date_ordered, date_status_changed,
			seller_name, store_name, buyer_name, buyer_email, buyer_order_count,
			status, is_invoiced, require_insurance, remarks, total_count, unique_count,
			total_weight, is_filed, drive_thru_sent, payment_method, payment_currency_code,
			payment_status, date_paid, shipping_method_id, shipping_method, tracking_link,
			shipping_address_name, shipping_address_full, shipping_country_code,
			currency_code, subtotal, total_price, etc1, etc2, insurance_cost,
			shipping_cost, credit_amount, coupon_amount, vat_rate, vat_amount,
			display_currency_code, display_subtotal, display_grand_total,
			display_etc1, display_etc2, display_insurance, display_shipping,
			display_credit, display_coupon, display_vat_rate, display_vat_amount,
			created_at, updated_at;`

	newOrder, err := db.QueryRowToStructFromService[bricklink.Order](repo.db, sql,
		userID, blOrder.OrderID, dateOrdered, dateStatusChanged, blOrder.SellerName, blOrder.StoreName,
		blOrder.BuyerName, blOrder.BuyerEmail, blOrder.BuyerOrderCount, blOrder.Status,
		blOrder.IsInvoiced, blOrder.RequireInsurance, blOrder.Remarks, blOrder.TotalCount,
		blOrder.UniqueCount, blOrder.TotalWeight, blOrder.IsFiled, blOrder.DriveThruSent,
		paymentMethod, paymentCurrencyCode, paymentStatus, datePaid,
		shippingMethodID, shippingMethod, trackingLink, shippingAddressName,
		shippingAddressFull, shippingCountryCode, currencyCode, subtotal,
		totalPrice, etc1, etc2, insuranceCost, shippingCost, creditAmount,
		couponAmount, vatRate, vatAmount, displayCurrencyCode, displaySubtotal,
		displayGrandTotal, displayEtc1, displayEtc2, displayInsurance,
		displayShipping, displayCredit, displayCoupon, displayVatRate,
		displayVatAmount)
	if err != nil {
		logger.Error("Database Error while creating order", "bricklink_order_id", blOrder.OrderID, "user_id", userID, "error", err)
		return nil, err
	}

	logger.Info("Created new order", "bricklink_order_id", blOrder.OrderID, "user_id", userID)
	return &OrderActionResult{Order: &newOrder, Created: true}, nil
}

// GetLastSyncTime retrieves the last sync time for a specific user
func (repo *OrderRepository) GetLastSyncTime(userID int64) (*time.Time, error) {
	sql := `SELECT last_sync_time 
			FROM order_syncs 
			WHERE user_id = $1 AND sync_status = 'completed'
			ORDER BY last_sync_time DESC 
			LIMIT 1;`

	rows, err := repo.db.CollectRowsToMap(sql, userID)
	if err != nil {
		logger.Error("Database Error while retrieving last sync time", "user_id", userID, "error", err)
		return nil, err
	}

	if len(rows) == 0 {
		return nil, nil // No previous sync
	}

	if syncTime, ok := rows[0]["last_sync_time"]; ok {
		if syncTimeVal, ok := syncTime.(time.Time); ok {
			return &syncTimeVal, nil
		}
	}

	return nil, nil
}

// CreateOrderSync creates a new order sync record
func (repo *OrderRepository) CreateOrderSync(userID int64, status string, ordersCount int, errorMessage *string) (*bricklink.OrderSync, error) {
	sql := `INSERT INTO order_syncs (user_id, last_sync_time, sync_status, orders_count, error_message, created_at) 
			VALUES ($1, CURRENT_TIMESTAMP, $2, $3, $4, CURRENT_TIMESTAMP) 
			RETURNING id, user_id, last_sync_time, sync_status, orders_count, error_message, created_at, updated_at;`

	orderSync, err := db.QueryRowToStructFromService[bricklink.OrderSync](repo.db, sql, userID, status, ordersCount, errorMessage)
	if err != nil {
		logger.Error("Database Error while creating order sync", "user_id", userID, "status", status, "error", err)
		return nil, err
	}

	logger.Info("Created order sync record", "user_id", userID, "status", status, "orders_count", ordersCount)
	return &orderSync, nil
}

// UpdateOrderSync updates an existing order sync record
func (repo *OrderRepository) UpdateOrderSync(syncID int64, status string, ordersCount int, errorMessage *string) (*bricklink.OrderSync, error) {
	sql := `UPDATE order_syncs 
			SET sync_status = $1, orders_count = $2, error_message = $3, updated_at = CURRENT_TIMESTAMP
			WHERE id = $4
			RETURNING id, user_id, last_sync_time, sync_status, orders_count, error_message, created_at, updated_at;`

	orderSync, err := db.QueryRowToStructFromService[bricklink.OrderSync](repo.db, sql, status, ordersCount, errorMessage, syncID)
	if err != nil {
		logger.Error("Database Error while updating order sync", "sync_id", syncID, "status", status, "error", err)
		return nil, err
	}

	logger.Info("Updated order sync record", "sync_id", syncID, "status", status, "orders_count", ordersCount)
	return &orderSync, nil
}

// GetOrdersCount returns the total count of orders for a user
func (repo *OrderRepository) GetOrdersCount(userID int64) (int64, error) {
	sql := `SELECT COUNT(*) as count FROM orders WHERE user_id = $1;`

	rows, err := repo.db.CollectRowsToMap(sql, userID)
	if err != nil {
		logger.Error("Database Error while counting orders", "user_id", userID, "error", err)
		return 0, err
	}

	if len(rows) > 0 {
		if count, ok := rows[0]["count"]; ok {
			if countVal, ok := count.(int64); ok {
				return countVal, nil
			}
		}
	}

	return 0, nil
}
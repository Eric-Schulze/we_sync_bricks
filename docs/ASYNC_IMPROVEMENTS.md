# Async Operations - Future TODO List

This document outlines areas in the codebase that could benefit from asynchronous operations to improve scalability and user experience.

## High Priority - Notification Processing

### 1. BrickLink Push Notification Processing
**Current:** Synchronous order sync when webhook received
**Location:** `internal/common/services/bricklink/bricklink_notifications.go`
**Issue:** Webhook handlers should respond quickly to BrickLink
**Solution:** Queue order sync operations
```go
// Instead of:
err = ordersService.SyncOrdersFromBrickLink(user)

// Do:
err = jobQueue.Enqueue("sync_orders", user.ID, "bricklink")
```

### 2. Order Synchronization
**Current:** Synchronous API calls to BrickLink/BrickOwl
**Location:** `orders/orders_service.go:SyncOrdersFromBrickLink()`
**Issue:** Can take several seconds, blocks UI refresh requests
**Solution:** Background job processing with progress updates
```go
// Background job that updates sync status in real-time
// UI can poll for progress or use WebSockets
```

### 3. Large Order Imports
**Current:** Bulk order processing in single request
**Location:** `orders/repository.go:CreateOrUpdateOrder()`
**Issue:** Importing hundreds of orders blocks other operations
**Solution:** Batch processing with job queue

## Medium Priority - User Experience

### 4. Minifig Parts Fetching
**Current:** Sequential API calls for parts data
**Location:** `partial_minifigs/service.go:GetMinifigParts()`
**Issue:** User waits for each part's pricing/pictures
**Solution:** Concurrent fetching with goroutines
```go
// Fetch parts data concurrently
var wg sync.WaitGroup
partsChan := make(chan PartData, len(parts))
```

### 5. Profile Updates with API Credential Validation
**Current:** Synchronous validation of BrickLink/BrickOwl credentials
**Location:** `internal/common/services/user/service.go:UpdateAPICredentials()`
**Issue:** Testing API credentials can take 5-10 seconds
**Solution:** Save credentials immediately, validate in background

### 6. Dashboard Data Loading
**Current:** Sequential queries for dashboard stats
**Location:** `internal/common/services/dashboard/service.go`
**Issue:** Multiple database queries executed serially
**Solution:** Concurrent query execution

## Low Priority - Performance Optimization

### 7. Image/Asset Caching
**Current:** Real-time fetching of part/minifig images
**Location:** BrickLink API image requests
**Issue:** Repeated API calls for same images
**Solution:** Background image pre-fetching and caching

### 8. Search Operations
**Current:** Synchronous database searches
**Location:** Various search endpoints
**Issue:** Complex searches can block other operations
**Solution:** Search result caching with background refresh

### 9. Backup Polling for Missed Notifications
**Current:** Not implemented yet
**Location:** To be implemented
**Issue:** Should run continuously without blocking other operations
**Solution:** Background cron job or scheduled worker

## Technical Implementation Patterns

### Job Queue Implementation Options
1. **Redis + Job Library** (e.g., `github.com/hibiken/asynq`)
2. **Database-based queue** (PostgreSQL with LISTEN/NOTIFY)
3. **Go channels with worker pools** (for simple cases)

### Progress Tracking
- Use WebSockets or Server-Sent Events for real-time updates
- Store job progress in Redis or database
- Provide job status endpoints for polling

### Error Handling
- Implement retry logic with exponential backoff
- Dead letter queues for failed jobs
- Comprehensive logging for debugging

### Monitoring
- Job queue metrics (pending, processing, failed)
- Processing time histograms
- Error rate monitoring

## Implementation Priority

1. **Start with notification processing** - Critical for webhook reliability
2. **Add order sync background jobs** - Improves user experience significantly
3. **Implement concurrent part fetching** - Quick win for user experience
4. **Add remaining optimizations** - As load increases

## Notes

- All async operations should have proper cancellation support (context.Context)
- Consider rate limiting for external API calls
- Implement circuit breakers for unreliable external services
- Always provide fallback/degraded functionality when background jobs fail
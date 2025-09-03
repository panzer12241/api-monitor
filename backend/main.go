package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
)

type APIEndpoint struct {
	ID                   int               `json:"id" db:"id"`
	Name                 string            `json:"name" db:"name"`
	URL                  string            `json:"url" db:"url"`
	Method               string            `json:"method" db:"method"`
	Headers              map[string]string `json:"headers" db:"headers"`
	Body                 string            `json:"body" db:"body"`
	TimeoutSeconds       int               `json:"timeout_seconds" db:"timeout_seconds"`
	CheckIntervalSeconds int               `json:"check_interval_seconds" db:"check_interval_seconds"`
	IsActive             bool              `json:"is_active" db:"is_active"`
	ProxyID              *int              `json:"proxy_id" db:"proxy_id"`
	Proxy                *Proxy            `json:"proxy,omitempty"`
	CreatedAt            time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at" db:"updated_at"`
}

type APICheckLog struct {
	ID              int       `json:"id"`
	EndpointID      int       `json:"endpoint_id"`
	StatusCode      int       `json:"status_code"`
	ResponseTimeMs  int       `json:"response_time_ms"`
	ResponseBody    string    `json:"response_body"`
	ResponseHeaders string    `json:"response_headers"`
	ErrorMessage    string    `json:"error_message"`
	CheckedAt       time.Time `json:"checked_at"`
}

type Proxy struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Host      string    `json:"host" db:"host"`
	Port      int       `json:"port" db:"port"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"password" db:"password"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Monitor struct {
	db         *sql.DB
	cron       *cron.Cron
	activeJobs map[int]cron.EntryID
	jobMutex   sync.RWMutex
}

func main() {
	// Connect to database
	db, err := connectDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize monitor
	monitor := &Monitor{
		db:         db,
		cron:       cron.New(),
		activeJobs: make(map[int]cron.EntryID),
	}

	// Start cron scheduler
	monitor.cron.Start()

	// Load existing active endpoints
	monitor.loadActiveEndpoints()

	// Schedule daily cleanup of old logs (run at 2 AM)
	monitor.cron.AddFunc("0 2 * * *", func() {
		monitor.cleanupOldLogs()
	})

	// Run initial cleanup
	go monitor.cleanupOldLogs()

	// Setup Gin router
	r := gin.Default()

	// No CORS middleware - let nginx handle CORS
	// r.Use(func(c *gin.Context) { ... })

	// Authentication endpoint (no auth required)
	r.POST("/api/v1/auth/login", monitor.login)

	// API endpoints
	api := r.Group("/api/v1")
	{
		// Endpoint management
		api.GET("/endpoints", monitor.getEndpoints)
		api.POST("/endpoints", monitor.createEndpoint)
		api.PUT("/endpoints/:id", monitor.updateEndpoint)
		api.DELETE("/endpoints/:id", monitor.deleteEndpoint)
		api.POST("/endpoints/:id/toggle", monitor.toggleEndpoint)
		api.GET("/endpoints/:id/logs", monitor.getEndpointLogs)
		api.POST("/endpoints/:id/check", monitor.manualCheck)
		api.POST("/cleanup-logs", monitor.manualCleanup)

		// Proxy management
		api.GET("/proxies", monitor.getProxies)
		api.POST("/proxies", monitor.createProxy)
		api.PUT("/proxies/:id", monitor.updateProxy)
		api.DELETE("/proxies/:id", monitor.deleteProxy)
		api.POST("/proxies/:id/toggle", monitor.toggleProxy)
	}

	log.Println("API Monitor started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func connectDB() (*sql.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres123")
	dbname := getEnv("DB_NAME", "api_monitor")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func (m *Monitor) loadActiveEndpoints() {
	rows, err := m.db.Query(`
		SELECT e.id, e.name, e.url, e.method, COALESCE(e.headers, '{}'), COALESCE(e.body, ''), 
		       e.timeout_seconds, e.check_interval_seconds, e.proxy_id,
		       p.host, p.port, p.username, p.password
		FROM api_endpoints e
		LEFT JOIN proxies p ON e.proxy_id = p.id AND p.is_active = true
		WHERE e.is_active = true`)
	if err != nil {
		log.Printf("Error loading active endpoints: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var endpoint APIEndpoint
		var proxy Proxy
		var headersJSON string
		var proxyID sql.NullInt64
		var proxyHost, proxyUsername, proxyPassword sql.NullString
		var proxyPort sql.NullInt64

		err := rows.Scan(&endpoint.ID, &endpoint.Name, &endpoint.URL, &endpoint.Method,
			&headersJSON, &endpoint.Body, &endpoint.TimeoutSeconds, &endpoint.CheckIntervalSeconds, &proxyID,
			&proxyHost, &proxyPort, &proxyUsername, &proxyPassword)
		if err != nil {
			log.Printf("Error scanning endpoint: %v", err)
			continue
		}

		json.Unmarshal([]byte(headersJSON), &endpoint.Headers)

		// Set proxy data if available
		if proxyID.Valid {
			endpoint.ProxyID = &[]int{int(proxyID.Int64)}[0]
			if proxyHost.Valid {
				proxy.Host = proxyHost.String
				proxy.Port = int(proxyPort.Int64)
				proxy.Username = proxyUsername.String
				proxy.Password = proxyPassword.String
				endpoint.Proxy = &proxy
			}
		}

		m.scheduleEndpoint(endpoint)
	}
}

func (m *Monitor) scheduleEndpoint(endpoint APIEndpoint) {
	m.jobMutex.Lock()
	defer m.jobMutex.Unlock()

	// Remove existing job if any
	if entryID, exists := m.activeJobs[endpoint.ID]; exists {
		m.cron.Remove(entryID)
	}

	// Schedule new job
	spec := fmt.Sprintf("@every %ds", endpoint.CheckIntervalSeconds)
	entryID, err := m.cron.AddFunc(spec, func() {
		m.checkEndpoint(endpoint)
	})

	if err != nil {
		log.Printf("Error scheduling endpoint %s: %v", endpoint.Name, err)
		return
	}

	m.activeJobs[endpoint.ID] = entryID
	log.Printf("Scheduled endpoint %s to check every %d seconds", endpoint.Name, endpoint.CheckIntervalSeconds)
}

func (m *Monitor) unscheduleEndpoint(endpointID int) {
	m.jobMutex.Lock()
	defer m.jobMutex.Unlock()

	if entryID, exists := m.activeJobs[endpointID]; exists {
		m.cron.Remove(entryID)
		delete(m.activeJobs, endpointID)
	}
}

func (m *Monitor) checkEndpoint(endpoint APIEndpoint) {
	start := time.Now()

	// Create HTTP client with optional proxy
	client := &http.Client{
		Timeout: time.Duration(endpoint.TimeoutSeconds) * time.Second,
	}

	// Configure proxy if specified
	if endpoint.Proxy != nil && endpoint.Proxy.Host != "" {
		proxyURL := fmt.Sprintf("http://%s:%d", endpoint.Proxy.Host, endpoint.Proxy.Port)

		// Add authentication if provided
		if endpoint.Proxy.Username != "" && endpoint.Proxy.Password != "" {
			proxyURL = fmt.Sprintf("http://%s:%s@%s:%d",
				endpoint.Proxy.Username, endpoint.Proxy.Password,
				endpoint.Proxy.Host, endpoint.Proxy.Port)
		}

		proxy, err := url.Parse(proxyURL)
		if err != nil {
			log.Printf("Error parsing proxy URL for endpoint %s: %v", endpoint.Name, err)
		} else {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}

	var req *http.Request
	var err error

	if endpoint.Body != "" {
		req, err = http.NewRequest(endpoint.Method, endpoint.URL, strings.NewReader(endpoint.Body))
	} else {
		req, err = http.NewRequest(endpoint.Method, endpoint.URL, nil)
	}

	if err != nil {
		m.logCheck(endpoint.ID, 0, 0, "", "", fmt.Sprintf("Error creating request: %v", err))
		return
	}

	// Add headers
	for key, value := range endpoint.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	duration := time.Since(start)
	durationMs := int(duration.Milliseconds())

	if err != nil {
		m.logCheck(endpoint.ID, 0, durationMs, "", "", fmt.Sprintf("Request failed: %v", err))
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Collect response headers
	headersStr := ""
	if resp.Header != nil {
		headers := make(map[string]string)
		for key, values := range resp.Header {
			if len(values) > 0 {
				headers[key] = values[0] // Take first value if multiple
			}
		}
		if headersBytes, err := json.Marshal(headers); err == nil {
			headersStr = string(headersBytes)
		}
	}

	// Truncate response body if too long
	if len(bodyStr) > 1000 {
		bodyStr = bodyStr[:1000] + "... (truncated)"
	}

	m.logCheck(endpoint.ID, resp.StatusCode, durationMs, bodyStr, headersStr, "")

	log.Printf("Checked %s: %d (%dms)", endpoint.Name, resp.StatusCode, durationMs)
}

func (m *Monitor) logCheck(endpointID, statusCode, responseTimeMs int, responseBody, responseHeaders, errorMessage string) {
	// Clean strings to ensure UTF-8 compatibility
	cleanResponseBody := strings.ToValidUTF8(responseBody, "")
	cleanResponseHeaders := strings.ToValidUTF8(responseHeaders, "")
	cleanErrorMessage := strings.ToValidUTF8(errorMessage, "")

	// Limit response body size to prevent database issues
	if len(cleanResponseBody) > 1000 {
		cleanResponseBody = cleanResponseBody[:1000] + "..."
	}

	// Limit response headers size
	if len(cleanResponseHeaders) > 2000 {
		cleanResponseHeaders = cleanResponseHeaders[:2000] + "..."
	}

	_, err := m.db.Exec(`
		INSERT INTO api_check_logs (endpoint_id, status_code, response_time_ms, response_body, response_headers, error_message)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		endpointID, statusCode, responseTimeMs, cleanResponseBody, cleanResponseHeaders, cleanErrorMessage)

	if err != nil {
		log.Printf("Error logging check: %v", err)
	}
}

func (m *Monitor) cleanupOldLogs() {
	// Delete logs older than 30 days
	result, err := m.db.Exec(`
		DELETE FROM api_check_logs 
		WHERE checked_at < NOW() - INTERVAL '30 days'`)

	if err != nil {
		log.Printf("Error cleaning up old logs: %v", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		log.Printf("Cleaned up %d logs older than 30 days", rowsAffected)
	}
}

func (m *Monitor) manualCleanup(c *gin.Context) {
	m.cleanupOldLogs()
	c.JSON(200, gin.H{"message": "Log cleanup completed"})
} // REST API Handlers

func (m *Monitor) getEndpoints(c *gin.Context) {
	rows, err := m.db.Query(`
		SELECT 
			e.id, e.name, e.url, e.method, 
			COALESCE(e.headers, '{}'), COALESCE(e.body, ''), 
			e.timeout_seconds, e.check_interval_seconds, e.is_active, 
			e.proxy_id, e.created_at, e.updated_at,
			p.id, p.name, p.host, p.port, p.username, p.password, p.is_active, p.created_at, p.updated_at
		FROM api_endpoints e
		LEFT JOIN proxies p ON e.proxy_id = p.id
		ORDER BY e.created_at DESC`)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var endpoints []APIEndpoint
	for rows.Next() {
		var endpoint APIEndpoint
		var proxy Proxy
		var headersJSON string
		var proxyID, proxyIdFromJoin sql.NullInt64
		var proxyName, proxyHost, proxyUsername, proxyPassword sql.NullString
		var proxyPort sql.NullInt64
		var proxyIsActive sql.NullBool
		var proxyCreatedAt, proxyUpdatedAt sql.NullTime

		err := rows.Scan(&endpoint.ID, &endpoint.Name, &endpoint.URL, &endpoint.Method,
			&headersJSON, &endpoint.Body, &endpoint.TimeoutSeconds, &endpoint.CheckIntervalSeconds,
			&endpoint.IsActive, &proxyID, &endpoint.CreatedAt, &endpoint.UpdatedAt,
			&proxyIdFromJoin, &proxyName, &proxyHost, &proxyPort, &proxyUsername, &proxyPassword,
			&proxyIsActive, &proxyCreatedAt, &proxyUpdatedAt)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		json.Unmarshal([]byte(headersJSON), &endpoint.Headers)

		// Set proxy_id
		if proxyID.Valid {
			endpoint.ProxyID = &[]int{int(proxyID.Int64)}[0]
		}

		// Set proxy object if exists
		if proxyIdFromJoin.Valid {
			proxy.ID = int(proxyIdFromJoin.Int64)
			proxy.Name = proxyName.String
			proxy.Host = proxyHost.String
			proxy.Port = int(proxyPort.Int64)
			proxy.Username = proxyUsername.String
			proxy.Password = proxyPassword.String
			proxy.IsActive = proxyIsActive.Bool
			proxy.CreatedAt = proxyCreatedAt.Time
			proxy.UpdatedAt = proxyUpdatedAt.Time
			endpoint.Proxy = &proxy
		}

		endpoints = append(endpoints, endpoint)
	}

	c.JSON(200, endpoints)
}

func (m *Monitor) createEndpoint(c *gin.Context) {
	var endpoint APIEndpoint
	if err := c.ShouldBindJSON(&endpoint); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if endpoint.Method == "" {
		endpoint.Method = "GET"
	}
	if endpoint.TimeoutSeconds == 0 {
		endpoint.TimeoutSeconds = 30
	}
	if endpoint.CheckIntervalSeconds == 0 {
		endpoint.CheckIntervalSeconds = 60
	}

	headersJSON, _ := json.Marshal(endpoint.Headers)

	err := m.db.QueryRow(`
		INSERT INTO api_endpoints (name, url, method, headers, body, timeout_seconds, check_interval_seconds, is_active, proxy_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`,
		endpoint.Name, endpoint.URL, endpoint.Method, string(headersJSON), endpoint.Body,
		endpoint.TimeoutSeconds, endpoint.CheckIntervalSeconds, endpoint.IsActive, endpoint.ProxyID).
		Scan(&endpoint.ID, &endpoint.CreatedAt, &endpoint.UpdatedAt)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if endpoint.IsActive {
		m.scheduleEndpoint(endpoint)
	}

	c.JSON(201, endpoint)
}

func (m *Monitor) updateEndpoint(c *gin.Context) {
	id := c.Param("id")
	endpointID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid endpoint ID"})
		return
	}

	var endpoint APIEndpoint
	if err := c.ShouldBindJSON(&endpoint); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	headersJSON, _ := json.Marshal(endpoint.Headers)

	_, err = m.db.Exec(`
		UPDATE api_endpoints 
		SET name = $1, url = $2, method = $3, headers = $4, body = $5, 
		    timeout_seconds = $6, check_interval_seconds = $7, is_active = $8, proxy_id = $9, updated_at = CURRENT_TIMESTAMP
		WHERE id = $10`,
		endpoint.Name, endpoint.URL, endpoint.Method, string(headersJSON), endpoint.Body,
		endpoint.TimeoutSeconds, endpoint.CheckIntervalSeconds, endpoint.IsActive, endpoint.ProxyID, endpointID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Update scheduling
	m.unscheduleEndpoint(endpointID)
	if endpoint.IsActive {
		// Reload endpoint with proxy data from database
		var updatedEndpoint APIEndpoint
		var proxy Proxy
		var headersJSON string
		var proxyID sql.NullInt64
		var proxyHost, proxyUsername, proxyPassword sql.NullString
		var proxyPort sql.NullInt64

		err = m.db.QueryRow(`
			SELECT e.id, e.name, e.url, e.method, COALESCE(e.headers, '{}'), COALESCE(e.body, ''), 
			       e.timeout_seconds, e.check_interval_seconds, e.proxy_id,
			       p.host, p.port, p.username, p.password
			FROM api_endpoints e
			LEFT JOIN proxies p ON e.proxy_id = p.id AND p.is_active = true
			WHERE e.id = $1`, endpointID).
			Scan(&updatedEndpoint.ID, &updatedEndpoint.Name, &updatedEndpoint.URL, &updatedEndpoint.Method,
				&headersJSON, &updatedEndpoint.Body, &updatedEndpoint.TimeoutSeconds,
				&updatedEndpoint.CheckIntervalSeconds, &proxyID,
				&proxyHost, &proxyPort, &proxyUsername, &proxyPassword)

		if err == nil {
			json.Unmarshal([]byte(headersJSON), &updatedEndpoint.Headers)

			// Set proxy data if available
			if proxyID.Valid {
				updatedEndpoint.ProxyID = &[]int{int(proxyID.Int64)}[0]
				if proxyHost.Valid {
					proxy.Host = proxyHost.String
					proxy.Port = int(proxyPort.Int64)
					proxy.Username = proxyUsername.String
					proxy.Password = proxyPassword.String
					updatedEndpoint.Proxy = &proxy
				}
			}

			m.scheduleEndpoint(updatedEndpoint)
		} else {
			log.Printf("Error reloading endpoint for scheduling: %v", err)
		}
	}

	c.JSON(200, gin.H{"message": "Endpoint updated successfully"})
}

func (m *Monitor) deleteEndpoint(c *gin.Context) {
	id := c.Param("id")
	endpointID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid endpoint ID"})
		return
	}

	// Get endpoint info before deleting
	var endpoint APIEndpoint
	err = m.db.QueryRow("SELECT id, name, url, method, timeout_seconds, check_interval_seconds, is_active, created_at, updated_at FROM api_endpoints WHERE id = $1", endpointID).Scan(
		&endpoint.ID, &endpoint.Name, &endpoint.URL, &endpoint.Method,
		&endpoint.TimeoutSeconds, &endpoint.CheckIntervalSeconds, &endpoint.IsActive,
		&endpoint.CreatedAt, &endpoint.UpdatedAt)
	if err != nil {
		c.JSON(404, gin.H{"error": "Endpoint not found"})
		return
	}

	m.unscheduleEndpoint(endpointID)

	// Delete endpoint logs first (foreign key constraint)
	_, err = m.db.Exec("DELETE FROM api_check_logs WHERE endpoint_id = $1", endpointID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete endpoint logs: " + err.Error()})
		return
	}

	// Delete endpoint
	_, err = m.db.Exec("DELETE FROM api_endpoints WHERE id = $1", endpointID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete endpoint: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Endpoint and all related logs deleted successfully"})
}

func (m *Monitor) toggleEndpoint(c *gin.Context) {
	id := c.Param("id")
	endpointID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid endpoint ID"})
		return
	}

	var isActive bool
	err = m.db.QueryRow("UPDATE api_endpoints SET is_active = NOT is_active, updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING is_active", endpointID).Scan(&isActive)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if isActive {
		// Load endpoint and schedule
		var endpoint APIEndpoint
		var headersJSON string
		err = m.db.QueryRow("SELECT id, name, url, method, COALESCE(headers, '{}'), COALESCE(body, ''), timeout_seconds, check_interval_seconds FROM api_endpoints WHERE id = $1", endpointID).
			Scan(&endpoint.ID, &endpoint.Name, &endpoint.URL, &endpoint.Method, &headersJSON, &endpoint.Body, &endpoint.TimeoutSeconds, &endpoint.CheckIntervalSeconds)
		if err == nil {
			json.Unmarshal([]byte(headersJSON), &endpoint.Headers)
			m.scheduleEndpoint(endpoint)
		}
	} else {
		m.unscheduleEndpoint(endpointID)
	}

	c.JSON(200, gin.H{"is_active": isActive})
}

func (m *Monitor) getEndpointLogs(c *gin.Context) {
	id := c.Param("id")
	endpointID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid endpoint ID"})
		return
	}

	limit := c.DefaultQuery("limit", "25")
	offset := c.DefaultQuery("offset", "0")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	minResponseTime := c.Query("min_response_time")
	statusCode := c.Query("status_code")

	// Debug logging
	log.Printf("Getting logs for endpoint %d: limit=%s, offset=%s, start_date=%s, end_date=%s",
		endpointID, limit, offset, startDate, endDate)

	// Build WHERE clause dynamically
	whereClause := "WHERE endpoint_id = $1"
	args := []interface{}{endpointID}
	argIndex := 2

	// Add date filters with validation
	if startDate != "" {
		// Try to parse the date to validate format
		if _, err := time.Parse(time.RFC3339, startDate); err != nil {
			c.JSON(400, gin.H{"error": "Invalid start_date format. Use ISO 8601 format (e.g., 2023-01-01T00:00:00Z)"})
			return
		}
		whereClause += fmt.Sprintf(" AND checked_at >= $%d", argIndex)
		args = append(args, startDate)
		argIndex++
	}
	if endDate != "" {
		// Try to parse the date to validate format
		if _, err := time.Parse(time.RFC3339, endDate); err != nil {
			c.JSON(400, gin.H{"error": "Invalid end_date format. Use ISO 8601 format (e.g., 2023-01-01T23:59:59Z)"})
			return
		}
		whereClause += fmt.Sprintf(" AND checked_at <= $%d", argIndex)
		args = append(args, endDate)
		argIndex++
	}

	// Add response time filter
	if minResponseTime != "" {
		whereClause += fmt.Sprintf(" AND response_time_ms >= $%d", argIndex)
		args = append(args, minResponseTime)
		argIndex++
	}

	// Add status code filter
	if statusCode != "" {
		if statusCode == "2xx" {
			whereClause += " AND status_code >= 200 AND status_code < 300"
		} else if statusCode == "3xx" {
			whereClause += " AND status_code >= 300 AND status_code < 400"
		} else if statusCode == "4xx" {
			whereClause += " AND status_code >= 400 AND status_code < 500"
		} else if statusCode == "5xx" {
			whereClause += " AND status_code >= 500 AND status_code < 600"
		} else {
			whereClause += fmt.Sprintf(" AND status_code = $%d", argIndex)
			args = append(args, statusCode)
			argIndex++
		}
	}

	// Get total count with filters
	countQuery := "SELECT COUNT(*) FROM api_check_logs " + whereClause
	var totalCount int
	err = m.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get total count: " + err.Error()})
		return
	}

	// Get paginated logs with filters - use optimized query
	query := fmt.Sprintf(`
		SELECT id, endpoint_id, status_code, response_time_ms, response_body, COALESCE(response_headers, ''), error_message, checked_at
		FROM api_check_logs 
		%s 
		ORDER BY checked_at DESC 
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	// Log the final query for debugging
	log.Printf("Executing query: %s with args: %v", query, args)

	rows, err := m.db.Query(query, args...)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var logs []APICheckLog
	for rows.Next() {
		var log APICheckLog
		err := rows.Scan(&log.ID, &log.EndpointID, &log.StatusCode, &log.ResponseTimeMs,
			&log.ResponseBody, &log.ResponseHeaders, &log.ErrorMessage, &log.CheckedAt)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		logs = append(logs, log)
	}

	// Return paginated response
	c.JSON(200, gin.H{
		"logs":       logs,
		"total":      totalCount,
		"limit":      limit,
		"offset":     offset,
		"start_date": startDate,
		"end_date":   endDate,
		"count":      len(logs),
	})
}

func (m *Monitor) manualCheck(c *gin.Context) {
	id := c.Param("id")
	endpointID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid endpoint ID"})
		return
	}

	// Get endpoint details
	var endpoint APIEndpoint
	var headersJSON string
	err = m.db.QueryRow(`
		SELECT id, name, url, method, COALESCE(headers, '{}'), COALESCE(body, ''), timeout_seconds, check_interval_seconds
		FROM api_endpoints WHERE id = $1`, endpointID).
		Scan(&endpoint.ID, &endpoint.Name, &endpoint.URL, &endpoint.Method,
			&headersJSON, &endpoint.Body, &endpoint.TimeoutSeconds, &endpoint.CheckIntervalSeconds)

	if err != nil {
		c.JSON(404, gin.H{"error": "Endpoint not found"})
		return
	}

	json.Unmarshal([]byte(headersJSON), &endpoint.Headers)

	// Perform check
	go m.checkEndpoint(endpoint)

	c.JSON(200, gin.H{"message": "Manual check initiated"})
}

// Login endpoint for authentication
func (m *Monitor) login(c *gin.Context) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	// Simple authentication (in production, use proper password hashing)
	if credentials.Username == "admin" && credentials.Password == "admin123" {
		token := fmt.Sprintf("demo-token-%d", time.Now().Unix())

		c.JSON(200, gin.H{
			"success": true,
			"token":   token,
			"user": gin.H{
				"id":       1,
				"username": credentials.Username,
				"name":     "Administrator",
			},
		})
	} else {
		c.JSON(401, gin.H{
			"success": false,
			"error":   "Invalid credentials",
		})
	}
}

// Proxy management methods
func (m *Monitor) getProxies(c *gin.Context) {
	rows, err := m.db.Query(`
		SELECT id, name, host, port, username, password, is_active, created_at, updated_at
		FROM proxies 
		ORDER BY created_at DESC`)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var proxies []Proxy
	for rows.Next() {
		var proxy Proxy
		err := rows.Scan(&proxy.ID, &proxy.Name, &proxy.Host, &proxy.Port,
			&proxy.Username, &proxy.Password, &proxy.IsActive,
			&proxy.CreatedAt, &proxy.UpdatedAt)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		proxies = append(proxies, proxy)
	}

	c.JSON(200, proxies)
}

func (m *Monitor) createProxy(c *gin.Context) {
	var proxy Proxy
	if err := c.ShouldBindJSON(&proxy); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if proxy.Name == "" || proxy.Host == "" || proxy.Port <= 0 {
		c.JSON(400, gin.H{"error": "Name, host, and port are required"})
		return
	}

	err := m.db.QueryRow(`
		INSERT INTO proxies (name, host, port, username, password, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, created_at, updated_at`,
		proxy.Name, proxy.Host, proxy.Port, proxy.Username, proxy.Password, proxy.IsActive).
		Scan(&proxy.ID, &proxy.CreatedAt, &proxy.UpdatedAt)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create proxy: " + err.Error()})
		return
	}

	c.JSON(201, proxy)
}

func (m *Monitor) updateProxy(c *gin.Context) {
	id := c.Param("id")
	proxyID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid proxy ID"})
		return
	}

	var proxy Proxy
	if err := c.ShouldBindJSON(&proxy); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	proxy.ID = proxyID

	err = m.db.QueryRow(`
		UPDATE proxies 
		SET name = $1, host = $2, port = $3, username = $4, password = $5, is_active = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING created_at, updated_at`,
		proxy.Name, proxy.Host, proxy.Port, proxy.Username, proxy.Password, proxy.IsActive, proxyID).
		Scan(&proxy.CreatedAt, &proxy.UpdatedAt)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update proxy: " + err.Error()})
		return
	}

	c.JSON(200, proxy)
}

func (m *Monitor) deleteProxy(c *gin.Context) {
	id := c.Param("id")
	proxyID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid proxy ID"})
		return
	}

	// Check if any endpoints are using this proxy
	var count int
	err = m.db.QueryRow("SELECT COUNT(*) FROM api_endpoints WHERE proxy_id = $1", proxyID).Scan(&count)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to check proxy usage: " + err.Error()})
		return
	}

	if count > 0 {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Cannot delete proxy: %d endpoints are using this proxy", count)})
		return
	}

	_, err = m.db.Exec("DELETE FROM proxies WHERE id = $1", proxyID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete proxy: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Proxy deleted successfully"})
}

func (m *Monitor) toggleProxy(c *gin.Context) {
	id := c.Param("id")
	proxyID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid proxy ID"})
		return
	}

	var isActive bool
	err = m.db.QueryRow("SELECT is_active FROM proxies WHERE id = $1", proxyID).Scan(&isActive)
	if err != nil {
		c.JSON(404, gin.H{"error": "Proxy not found"})
		return
	}

	newStatus := !isActive
	_, err = m.db.Exec("UPDATE proxies SET is_active = $1, updated_at = NOW() WHERE id = $2", newStatus, proxyID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to toggle proxy status: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"id":        proxyID,
		"is_active": newStatus,
		"message":   fmt.Sprintf("Proxy %s", map[bool]string{true: "activated", false: "deactivated"}[newStatus]),
	})
}

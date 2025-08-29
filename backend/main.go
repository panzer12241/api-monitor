package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

type Monitor struct {
	db              *sql.DB
	cron            *cron.Cron
	activeJobs      map[int]cron.EntryID
	jobMutex        sync.RWMutex
	prometheusGauge *prometheus.GaugeVec
	responseTime    *prometheus.HistogramVec
}

func main() {
	// Connect to database
	db, err := connectDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize Prometheus metrics
	prometheusGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "api_endpoint_status",
			Help: "Status of API endpoints (1 = up, 0 = down)",
		},
		[]string{"endpoint_id", "endpoint_name", "url", "method"},
	)

	responseTime := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_endpoint_response_time_seconds",
			Help:    "Response time of API endpoints in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint_id", "endpoint_name", "url", "method"},
	)

	prometheus.MustRegister(prometheusGauge)
	prometheus.MustRegister(responseTime)

	// Initialize monitor
	monitor := &Monitor{
		db:              db,
		cron:            cron.New(),
		activeJobs:      make(map[int]cron.EntryID),
		prometheusGauge: prometheusGauge,
		responseTime:    responseTime,
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

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Authentication endpoint (no auth required)
	r.POST("/api/v1/auth/login", monitor.login)

	// API endpoints
	api := r.Group("/api/v1")
	{
		api.GET("/endpoints", monitor.getEndpoints)
		api.POST("/endpoints", monitor.createEndpoint)
		api.PUT("/endpoints/:id", monitor.updateEndpoint)
		api.DELETE("/endpoints/:id", monitor.deleteEndpoint)
		api.POST("/endpoints/:id/toggle", monitor.toggleEndpoint)
		api.GET("/endpoints/:id/logs", monitor.getEndpointLogs)
		api.POST("/endpoints/:id/check", monitor.manualCheck)
		api.POST("/cleanup-logs", monitor.manualCleanup)
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
	rows, err := m.db.Query("SELECT id, name, url, method, COALESCE(headers, '{}'), COALESCE(body, ''), timeout_seconds, check_interval_seconds FROM api_endpoints WHERE is_active = true")
	if err != nil {
		log.Printf("Error loading active endpoints: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var endpoint APIEndpoint
		var headersJSON string

		err := rows.Scan(&endpoint.ID, &endpoint.Name, &endpoint.URL, &endpoint.Method,
			&headersJSON, &endpoint.Body, &endpoint.TimeoutSeconds, &endpoint.CheckIntervalSeconds)
		if err != nil {
			log.Printf("Error scanning endpoint: %v", err)
			continue
		}

		json.Unmarshal([]byte(headersJSON), &endpoint.Headers)
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

	client := &http.Client{
		Timeout: time.Duration(endpoint.TimeoutSeconds) * time.Second,
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
		m.updatePrometheusMetrics(endpoint, 0, 0)
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
		m.updatePrometheusMetrics(endpoint, 0, duration.Seconds())
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

	// Update Prometheus metrics
	status := 0.0
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		status = 1.0
	}
	m.updatePrometheusMetrics(endpoint, status, duration.Seconds())

	log.Printf("Checked %s: %d (%dms)", endpoint.Name, resp.StatusCode, durationMs)
}

func (m *Monitor) updatePrometheusMetrics(endpoint APIEndpoint, status float64, responseTime float64) {
	labels := prometheus.Labels{
		"endpoint_id":   strconv.Itoa(endpoint.ID),
		"endpoint_name": endpoint.Name,
		"url":           endpoint.URL,
		"method":        endpoint.Method,
	}

	m.prometheusGauge.With(labels).Set(status)
	m.responseTime.With(labels).Observe(responseTime)
}

func (m *Monitor) deletePrometheusMetrics(endpoint APIEndpoint) {
	labels := prometheus.Labels{
		"endpoint_id":   strconv.Itoa(endpoint.ID),
		"endpoint_name": endpoint.Name,
		"url":           endpoint.URL,
		"method":        endpoint.Method,
	}

	// Set status to 0 (down) and delete the metric
	m.prometheusGauge.With(labels).Set(0)
	m.prometheusGauge.Delete(labels)

	// Note: Histogram metrics cannot be completely deleted in Prometheus
	// but they will stop being updated and eventually be cleaned up by Prometheus
	m.responseTime.Delete(labels)
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
		SELECT id, name, url, method, COALESCE(headers, '{}'), COALESCE(body, ''), timeout_seconds, check_interval_seconds, is_active, created_at, updated_at
		FROM api_endpoints ORDER BY created_at DESC`)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var endpoints []APIEndpoint
	for rows.Next() {
		var endpoint APIEndpoint
		var headersJSON string

		err := rows.Scan(&endpoint.ID, &endpoint.Name, &endpoint.URL, &endpoint.Method,
			&headersJSON, &endpoint.Body, &endpoint.TimeoutSeconds, &endpoint.CheckIntervalSeconds,
			&endpoint.IsActive, &endpoint.CreatedAt, &endpoint.UpdatedAt)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		json.Unmarshal([]byte(headersJSON), &endpoint.Headers)
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
		INSERT INTO api_endpoints (name, url, method, headers, body, timeout_seconds, check_interval_seconds, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`,
		endpoint.Name, endpoint.URL, endpoint.Method, string(headersJSON), endpoint.Body,
		endpoint.TimeoutSeconds, endpoint.CheckIntervalSeconds, endpoint.IsActive).
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
		    timeout_seconds = $6, check_interval_seconds = $7, is_active = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $9`,
		endpoint.Name, endpoint.URL, endpoint.Method, string(headersJSON), endpoint.Body,
		endpoint.TimeoutSeconds, endpoint.CheckIntervalSeconds, endpoint.IsActive, endpointID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Update scheduling
	m.unscheduleEndpoint(endpointID)
	if endpoint.IsActive {
		endpoint.ID = endpointID
		m.scheduleEndpoint(endpoint)
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

	// Get endpoint info before deleting (for Prometheus cleanup)
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

	// Delete Prometheus metrics for this endpoint
	m.deletePrometheusMetrics(endpoint)

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

	limit := c.DefaultQuery("limit", "100")

	rows, err := m.db.Query(`
		SELECT id, endpoint_id, status_code, response_time_ms, response_body, COALESCE(response_headers, ''), error_message, checked_at
		FROM api_check_logs 
		WHERE endpoint_id = $1 
		ORDER BY checked_at DESC 
		LIMIT $2`, endpointID, limit)

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

	c.JSON(200, logs)
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

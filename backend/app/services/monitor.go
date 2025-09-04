package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"api-monitor/app/models"
	"api-monitor/utils"

	"github.com/robfig/cron/v3"
)

type MonitorService struct {
	DB         *sql.DB
	Cron       *cron.Cron
	ActiveJobs map[int]cron.EntryID
	JobMutex   sync.RWMutex
}

func NewMonitorService(db *sql.DB) *MonitorService {
	return &MonitorService{
		DB:         db,
		Cron:       cron.New(),
		ActiveJobs: make(map[int]cron.EntryID),
	}
}

func (m *MonitorService) Start() {
	m.Cron.Start()
	m.LoadActiveEndpoints()

	// Schedule daily cleanup of old logs (run at 2 AM)
	m.Cron.AddFunc("0 2 * * *", func() {
		m.CleanupOldLogs()
	})

	// Run initial cleanup
	go m.CleanupOldLogs()
}

func (m *MonitorService) Stop() {
	m.Cron.Stop()
}

func (m *MonitorService) LoadActiveEndpoints() {
	rows, err := m.DB.Query(`
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
		var endpoint models.APIEndpoint
		var proxy models.Proxy
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

		m.ScheduleEndpoint(endpoint)
	}
}

func (m *MonitorService) ScheduleEndpoint(endpoint models.APIEndpoint) {
	m.JobMutex.Lock()
	defer m.JobMutex.Unlock()

	// Remove existing job if any
	if entryID, exists := m.ActiveJobs[endpoint.ID]; exists {
		m.Cron.Remove(entryID)
	}

	// Schedule new job
	spec := fmt.Sprintf("@every %ds", endpoint.CheckIntervalSeconds)
	entryID, err := m.Cron.AddFunc(spec, func() {
		m.checkEndpoint(endpoint)
	})

	if err != nil {
		log.Printf("Error scheduling endpoint %s: %v", endpoint.Name, err)
		return
	}

	m.ActiveJobs[endpoint.ID] = entryID
	log.Printf("Scheduled endpoint %s to check every %d seconds", endpoint.Name, endpoint.CheckIntervalSeconds)
}

func (m *MonitorService) UnscheduleEndpoint(endpointID int) {
	m.JobMutex.Lock()
	defer m.JobMutex.Unlock()

	if entryID, exists := m.ActiveJobs[endpointID]; exists {
		m.Cron.Remove(entryID)
		delete(m.ActiveJobs, endpointID)
	}
}

func (m *MonitorService) checkEndpoint(endpoint models.APIEndpoint) {
	statusCode, responseTimeMs, responseBody, responseHeaders, err := utils.CheckEndpoint(endpoint)

	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
		log.Printf("Error checking endpoint %s: %v", endpoint.Name, err)
	} else {
		log.Printf("Checked %s: %d (%dms)", endpoint.Name, statusCode, responseTimeMs)
	}

	m.logCheck(endpoint.ID, statusCode, responseTimeMs, responseBody, responseHeaders, errorMessage)
}

func (m *MonitorService) logCheck(endpointID, statusCode, responseTimeMs int, responseBody, responseHeaders, errorMessage string) {
	// Clean strings to ensure UTF-8 compatibility
	cleanResponseBody := utils.ValidateUTF8(responseBody)
	cleanResponseHeaders := utils.ValidateUTF8(responseHeaders)
	cleanErrorMessage := utils.ValidateUTF8(errorMessage)

	// Limit response body size to prevent database issues
	if len(cleanResponseBody) > 1000 {
		cleanResponseBody = cleanResponseBody[:1000] + "..."
	}

	// Limit response headers size
	if len(cleanResponseHeaders) > 2000 {
		cleanResponseHeaders = cleanResponseHeaders[:2000] + "..."
	}

	_, err := m.DB.Exec(`
		INSERT INTO api_check_logs (endpoint_id, status_code, response_time_ms, response_body, response_headers, error_message)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		endpointID, statusCode, responseTimeMs, cleanResponseBody, cleanResponseHeaders, cleanErrorMessage)

	if err != nil {
		log.Printf("Error logging check: %v", err)
	}
}

func (m *MonitorService) CleanupOldLogs() {
	// Delete logs older than 30 days
	result, err := m.DB.Exec(`
		DELETE FROM api_check_logs 
		WHERE checked_at < NOW() - INTERVAL '30 days'`)

	if err != nil {
		log.Printf("Error cleaning up old logs: %v", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		log.Printf("Cleaned up %d logs older than 30 days", rowsAffected)
	}
}

package models

import (
	"time"
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

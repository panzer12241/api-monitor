package controllers

import (
	"database/sql"
	"encoding/json"
	"strconv"

	"api-monitor/app/models"
	"api-monitor/app/services"

	"github.com/gofiber/fiber/v2"
)

type EndpointController struct {
	DB      *sql.DB
	Monitor *services.MonitorService
}

func NewEndpointController(db *sql.DB, monitor *services.MonitorService) *EndpointController {
	return &EndpointController{
		DB:      db,
		Monitor: monitor,
	}
}

func (ec *EndpointController) GetEndpoints(c *fiber.Ctx) error {
	query := `
		SELECT id, name, url, method, headers, body, timeout_seconds, 
		       check_interval_seconds, is_active, proxy_id, created_at, updated_at
		FROM api_endpoints 
		ORDER BY created_at DESC
	`

	rows, err := ec.DB.Query(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch endpoints",
		})
	}
	defer rows.Close()

	var endpoints []models.APIEndpoint
	for rows.Next() {
		var endpoint models.APIEndpoint
		var proxyID sql.NullInt64
		var headersJSON sql.NullString

		err := rows.Scan(
			&endpoint.ID,
			&endpoint.Name,
			&endpoint.URL,
			&endpoint.Method,
			&headersJSON,
			&endpoint.Body,
			&endpoint.TimeoutSeconds,
			&endpoint.CheckIntervalSeconds,
			&endpoint.IsActive,
			&proxyID,
			&endpoint.CreatedAt,
			&endpoint.UpdatedAt,
		)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to scan endpoint data: " + err.Error(),
			})
		}

		// Handle headers JSON
		endpoint.Headers = make(map[string]string)
		if headersJSON.Valid && headersJSON.String != "" {
			// Try to parse JSON headers
			if err := json.Unmarshal([]byte(headersJSON.String), &endpoint.Headers); err != nil {
				// If JSON parsing fails, initialize empty map
				endpoint.Headers = make(map[string]string)
			}
		}

		// Handle proxy ID
		if proxyID.Valid {
			proxyIDInt := int(proxyID.Int64)
			endpoint.ProxyID = &proxyIDInt
		}

		endpoints = append(endpoints, endpoint)
	}

	return c.JSON(fiber.Map{
		"data": endpoints,
	})
}

func (ec *EndpointController) CreateEndpoint(c *fiber.Ctx) error {
	var endpoint models.APIEndpoint
	if err := c.BodyParser(&endpoint); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Validate required fields
	if endpoint.Name == "" || endpoint.URL == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name and URL are required",
		})
	}

	// Convert headers map to JSON string
	var headersJSON string
	if len(endpoint.Headers) > 0 {
		headersBytes, err := json.Marshal(endpoint.Headers)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid headers format",
			})
		}
		headersJSON = string(headersBytes)
	} else {
		headersJSON = "{}"
	}

	query := `
		INSERT INTO api_endpoints (name, url, method, headers, body, timeout_seconds, 
		                          check_interval_seconds, is_active, proxy_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	var proxyID interface{}
	if endpoint.ProxyID != nil {
		proxyID = *endpoint.ProxyID
	}

	err := ec.DB.QueryRow(
		query,
		endpoint.Name,
		endpoint.URL,
		endpoint.Method,
		headersJSON,
		endpoint.Body,
		endpoint.TimeoutSeconds,
		endpoint.CheckIntervalSeconds,
		endpoint.IsActive,
		proxyID,
	).Scan(&endpoint.ID, &endpoint.CreatedAt, &endpoint.UpdatedAt)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create endpoint: " + err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Endpoint created successfully",
		"data":    endpoint,
	})
}

func (ec *EndpointController) UpdateEndpoint(c *fiber.Ctx) error {
	id := c.Params("id")
	endpointID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid endpoint ID"})
	}

	var endpoint models.APIEndpoint
	if err := c.BodyParser(&endpoint); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Validate required fields
	if endpoint.Name == "" || endpoint.URL == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name and URL are required",
		})
	}

	query := `
		UPDATE api_endpoints 
		SET name = $1, url = $2, method = $3, headers = $4, body = $5, 
		    timeout_seconds = $6, check_interval_seconds = $7, is_active = $8, 
		    proxy_id = $9, updated_at = NOW()
		WHERE id = $10
		RETURNING id, created_at, updated_at
	`

	var proxyID interface{}
	if endpoint.ProxyID != nil {
		proxyID = *endpoint.ProxyID
	}

	err = ec.DB.QueryRow(
		query,
		endpoint.Name,
		endpoint.URL,
		endpoint.Method,
		endpoint.Headers,
		endpoint.Body,
		endpoint.TimeoutSeconds,
		endpoint.CheckIntervalSeconds,
		endpoint.IsActive,
		proxyID,
		endpointID,
	).Scan(&endpoint.ID, &endpoint.CreatedAt, &endpoint.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{
				"error": "Endpoint not found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update endpoint",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Endpoint updated successfully",
		"data":    endpoint,
	})
}

func (ec *EndpointController) DeleteEndpoint(c *fiber.Ctx) error {
	id := c.Params("id")
	endpointID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid endpoint ID"})
	}

	query := "DELETE FROM api_endpoints WHERE id = $1"
	result, err := ec.DB.Exec(query, endpointID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete endpoint",
		})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to verify deletion",
		})
	}

	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "Endpoint not found",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "Endpoint deleted successfully",
	})
}

func (ec *EndpointController) ToggleEndpoint(c *fiber.Ctx) error {
	id := c.Params("id")
	endpointID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid endpoint ID"})
	}

	query := `
		UPDATE api_endpoints 
		SET is_active = NOT is_active, updated_at = NOW()
		WHERE id = $1
		RETURNING is_active
	`

	var isActive bool
	err = ec.DB.QueryRow(query, endpointID).Scan(&isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{
				"error": "Endpoint not found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to toggle endpoint status",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "Endpoint status toggled successfully",
		"data": fiber.Map{
			"id":        endpointID,
			"is_active": isActive,
		},
	})
}

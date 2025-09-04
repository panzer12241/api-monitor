package controllers

import (
	"api-monitor/app/models"
	"database/sql"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ProxyController struct {
	db *sql.DB
}

func NewProxyController(db *sql.DB) *ProxyController {
	return &ProxyController{db: db}
}

// GetProxies retrieves all proxies
func (pc *ProxyController) GetProxies(c *fiber.Ctx) error {
	query := `
		SELECT id, name, host, port, username, password, is_active, created_at, updated_at 
		FROM proxies 
		ORDER BY created_at DESC
	`

	rows, err := pc.db.Query(query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch proxies",
		})
	}
	defer rows.Close()

	var proxies []models.Proxy
	for rows.Next() {
		var proxy models.Proxy
		err := rows.Scan(
			&proxy.ID,
			&proxy.Name,
			&proxy.Host,
			&proxy.Port,
			&proxy.Username,
			&proxy.Password,
			&proxy.IsActive,
			&proxy.CreatedAt,
			&proxy.UpdatedAt,
		)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to scan proxy data",
			})
		}
		proxies = append(proxies, proxy)
	}

	return c.JSON(fiber.Map{
		"data": proxies,
	})
}

// CreateProxy creates a new proxy
func (pc *ProxyController) CreateProxy(c *fiber.Ctx) error {
	var proxy models.Proxy
	if err := c.BodyParser(&proxy); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if proxy.Name == "" || proxy.Host == "" || proxy.Port == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name, host, and port are required",
		})
	}

	now := time.Now()
	query := `
		INSERT INTO proxies (name, host, port, username, password, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := pc.db.QueryRow(
		query,
		proxy.Name,
		proxy.Host,
		proxy.Port,
		proxy.Username,
		proxy.Password,
		proxy.IsActive,
		now,
		now,
	).Scan(&proxy.ID, &proxy.CreatedAt, &proxy.UpdatedAt)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create proxy",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Proxy created successfully",
		"data":    proxy,
	})
}

// UpdateProxy updates an existing proxy
func (pc *ProxyController) UpdateProxy(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid proxy ID",
		})
	}

	var proxy models.Proxy
	if err := c.BodyParser(&proxy); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if proxy.Name == "" || proxy.Host == "" || proxy.Port == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name, host, and port are required",
		})
	}

	query := `
		UPDATE proxies 
		SET name = $1, host = $2, port = $3, username = $4, password = $5, is_active = $6, updated_at = $7
		WHERE id = $8
		RETURNING id, created_at, updated_at
	`

	err = pc.db.QueryRow(
		query,
		proxy.Name,
		proxy.Host,
		proxy.Port,
		proxy.Username,
		proxy.Password,
		proxy.IsActive,
		time.Now(),
		id,
	).Scan(&proxy.ID, &proxy.CreatedAt, &proxy.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{
				"error": "Proxy not found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update proxy",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Proxy updated successfully",
		"data":    proxy,
	})
}

// DeleteProxy deletes a proxy
func (pc *ProxyController) DeleteProxy(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid proxy ID",
		})
	}

	query := "DELETE FROM proxies WHERE id = $1"
	result, err := pc.db.Exec(query, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete proxy",
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
			"error": "Proxy not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Proxy deleted successfully",
	})
}

// ToggleProxy toggles the active status of a proxy
func (pc *ProxyController) ToggleProxy(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid proxy ID",
		})
	}

	query := `
		UPDATE proxies 
		SET is_active = NOT is_active, updated_at = $1
		WHERE id = $2
		RETURNING is_active
	`

	var isActive bool
	err = pc.db.QueryRow(query, time.Now(), id).Scan(&isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{
				"error": "Proxy not found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to toggle proxy status",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Proxy status toggled successfully",
		"data": fiber.Map{
			"id":        id,
			"is_active": isActive,
		},
	})
}

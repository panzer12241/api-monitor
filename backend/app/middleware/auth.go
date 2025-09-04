package middleware

import (
	"fmt"
	"strings"

	"api-monitor/app/models"
	"api-monitor/config"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWT Middleware
func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{"error": "Authorization header required"})
		}

		// Extract token from Bearer scheme
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return c.Status(401).JSON(fiber.Map{"error": "Bearer token required"})
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return config.GetJWTSecret(), nil
		})

		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
		}

		if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
			// Check if token is blacklisted
			if isBlacklisted, err := checkTokenBlacklist(claims.ID); err != nil {
				return c.Status(500).JSON(fiber.Map{"error": "Authentication error"})
			} else if isBlacklisted {
				return c.Status(401).JSON(fiber.Map{"error": "Token has been revoked"})
			}

			// Create user object for context
			user := &models.User{
				ID:       claims.UserID,
				Username: claims.Username,
				Role:     claims.Role,
			}

			// Store user info in context
			c.Locals("user", user)
			c.Locals("userID", claims.UserID)
			c.Locals("username", claims.Username)
			c.Locals("role", claims.Role)
			return c.Next()
		}

		return c.Status(401).JSON(fiber.Map{"error": "Invalid token claims"})
	}
}

// Check if token is blacklisted
func checkTokenBlacklist(jti string) (bool, error) {
	db, err := config.ConnectDBWithoutMigration()
	if err != nil {
		return false, err
	}
	defer db.Close()

	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM blacklisted_tokens 
		WHERE token_jti = $1 AND expires_at > NOW()`, jti).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Admin Middleware
func AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role != "admin" {
			return c.Status(403).JSON(fiber.Map{"error": "Admin access required"})
		}
		return c.Next()
	}
}

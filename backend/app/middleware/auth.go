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
			// Store user info in context
			c.Locals("userID", claims.UserID)
			c.Locals("username", claims.Username)
			c.Locals("role", claims.Role)
			return c.Next()
		}

		return c.Status(401).JSON(fiber.Map{"error": "Invalid token claims"})
	}
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

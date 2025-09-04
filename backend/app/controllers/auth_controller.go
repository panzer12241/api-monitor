package controllers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log"
	"strings"
	"time"

	"api-monitor/app/models"
	"api-monitor/config"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	DB *sql.DB
}

func NewAuthController(db *sql.DB) *AuthController {
	return &AuthController{
		DB: db,
	}
}

// Login endpoint
func (ac *AuthController) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	log.Printf("Login attempt for username: %s", req.Username)

	// Get user from database
	var user models.User
	err := ac.DB.QueryRow(`
		SELECT id, username, password, role, is_active, created_at, updated_at
		FROM users WHERE username = $1 AND is_active = true`, req.Username).
		Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User not found: %s", req.Username)
			return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
		}
		log.Printf("Database error: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Database error"})
	}

	log.Printf("User found: %s, checking password...", user.Username)

	// Check password
	if !ac.checkPassword(req.Password, user.Password) {
		log.Printf("Password check failed for user: %s", user.Username)
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	log.Printf("Login successful for user: %s", user.Username)

	// Generate JWT token
	token, err := ac.generateJWT(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not generate token"})
	}

	// Return response without password
	user.Password = ""
	return c.JSON(models.LoginResponse{
		Token: token,
		User:  user,
	})
}

// Logout endpoint
func (ac *AuthController) Logout(c *fiber.Ctx) error {
	// Get user from context (set by middleware)
	user := c.Locals("user")
	if user == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	userObj, ok := user.(*models.User)
	if !ok {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid user data"})
	}

	// Get JWT token from header
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(401).JSON(fiber.Map{"error": "No token provided"})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Parse token to get JTI
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return config.GetJWTSecret(), nil
	})

	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
	}

	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid token claims"})
	}

	// Add token to blacklist
	_, err = ac.DB.Exec(`
		INSERT INTO blacklisted_tokens (token_jti, user_id, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (token_jti) DO NOTHING`,
		claims.ID, userObj.ID, claims.ExpiresAt.Time)

	if err != nil {
		log.Printf("Failed to blacklist token: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to logout"})
	}

	log.Printf("User logged out: %s (token blacklisted)", userObj.Username)

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
		"note":    "Token has been invalidated",
	})
}

// Register endpoint
func (ac *AuthController) Register(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Username and password are required"})
	}

	if req.Role == "" {
		req.Role = "user" // Default role
	}

	// Hash password
	hashedPassword, err := ac.hashPassword(req.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not hash password"})
	}

	// Insert user into database
	var user models.User
	err = ac.DB.QueryRow(`
		INSERT INTO users (username, password, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, true, NOW(), NOW())
		RETURNING id, username, role, is_active, created_at, updated_at`,
		req.Username, hashedPassword, req.Role).
		Scan(&user.ID, &user.Username, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return c.Status(409).JSON(fiber.Map{"error": "Username already exists"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Could not create user"})
	}

	// Generate JWT token
	token, err := ac.generateJWT(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not generate token"})
	}

	return c.Status(201).JSON(models.LoginResponse{
		Token: token,
		User:  user,
	})
}

// Generate JWT token
func (ac *AuthController) generateJWT(user models.User) (string, error) {
	// Generate unique JTI (JWT ID)
	jti, err := ac.generateJTI()
	if err != nil {
		return "", err
	}

	expirationTime := time.Now().Add(24 * time.Hour)

	claims := models.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(config.GetJWTSecret())
}

// Generate unique JTI (JWT ID)
func (ac *AuthController) generateJTI() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Hash password
func (ac *AuthController) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Check password
func (ac *AuthController) checkPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

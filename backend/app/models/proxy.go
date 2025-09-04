package models

import (
	"time"
)

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

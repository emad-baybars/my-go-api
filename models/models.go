package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gorm.io/gorm"
)

// User represents user model for PostgreSQL
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"`
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	Role      string         `json:"role" gorm:"default:user"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// UserMongo represents user model for MongoDB
type UserMongo struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email     string             `json:"email" bson:"email"`
	Username  string             `json:"username" bson:"username"`
	Password  string             `json:"-" bson:"password"`
	FirstName string             `json:"first_name" bson:"first_name"`
	LastName  string             `json:"last_name" bson:"last_name"`
	Role      string             `json:"role" bson:"role"`
	IsActive  bool               `json:"is_active" bson:"is_active"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// RegisterRequest represents registration request payload
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email" example:"user@example.com"`
	Username  string `json:"username" binding:"required,min=3" example:"username"`
	Password  string `json:"password" binding:"required,min=6" example:"password123"`
	FirstName string `json:"first_name" binding:"required" example:"John"`
	LastName  string `json:"last_name" binding:"required" example:"Doe"`
}

// UpdateUserRequest represents user update request payload
type UpdateUserRequest struct {
	FirstName string `json:"first_name" example:"John"`
	LastName  string `json:"last_name" example:"Doe"`
	Email     string `json:"email" binding:"omitempty,email" example:"user@example.com"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	User      UserInfo  `json:"user"`
	ExpiresAt time.Time `json:"expires_at" example:"2024-01-01T00:00:00Z"`
}

// UserInfo represents public user information
type UserInfo struct {
	ID        interface{} `json:"id"`
	Email     string      `json:"email" example:"user@example.com"`
	Username  string      `json:"username" example:"username"`
	FirstName string      `json:"first_name" example:"John"`
	LastName  string      `json:"last_name" example:"Doe"`
	Role      string      `json:"role" example:"user"`
	IsActive  bool        `json:"is_active" example:"true"`
	CreatedAt time.Time   `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time   `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// APIResponse represents standard API response
type APIResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Operation successful"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty" example:"Error message"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string            `json:"status" example:"healthy"`
	Timestamp time.Time         `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	Services  map[string]string `json:"services"`
	Version   string            `json:"version" example:"1.0.0"`
}

// PaginationQuery represents pagination query parameters
type PaginationQuery struct {
	Page     int    `form:"page,default=1" binding:"min=1" example:"1"`
	PageSize int    `form:"page_size,default=10" binding:"min=1,max=100" example:"10"`
	Sort     string `form:"sort" example:"created_at:desc"`
	Search   string `form:"search" example:"john"`
}

// PaginatedResponse represents paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page      int   `json:"page" example:"1"`
	PageSize  int   `json:"page_size" example:"10"`
	Total     int64 `json:"total" example:"100"`
	TotalPage int   `json:"total_page" example:"10"`
}

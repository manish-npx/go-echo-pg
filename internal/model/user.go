package model

import "github.com/jackc/pgx/v5/pgtype"

type User struct {
	ID        pgtype.UUID        `json:"id"`
	Email     string             `json:"email"`
	Password  string             `json:"-"`
	Name      string             `json:"name"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UpdateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

type AuthResponse struct {
	User      *User  `json:"user"`
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

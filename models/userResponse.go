package models

import "github.com/google/uuid"

type UserResponse struct {
	Message   string  `json:"message"`
	Name      string  `json:"name" db:"name"`
	Email     string  `json:"email" db:"email"`
	Address   string  `json:"address" db:"address"`
	Latitude  float64 `json:"late" db:"latitude"`
	Longitude float64 `json:"long" db:"longitude"`
}
type LoginResponse struct {
	Message               string    `json:"message"`
	Name                  string    `json:"name" db:"name"`
	UserSessionIDResponse uuid.UUID `json:"user_session_id"`
}
type UserWithRoleResponse struct {
	Message   string  `json:"message"`
	Name      string  `json:"name" db:"name"`
	Role      string  `json:"role" db:"role"`
	Email     string  `json:"email" db:"email"`
	Address   string  `json:"address" db:"address"`
	Latitude  float64 `json:"late" db:"latitude"`
	Longitude float64 `json:"long" db:"longitude"`
}
type UserSessionResponse struct {
	UserID  uuid.UUID `json:"userID" db:"user_id"`
	RolesID uuid.UUID `json:"rolesID" db:"roles_id"`
}

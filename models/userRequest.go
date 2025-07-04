package models

import "github.com/google/uuid"

type UserRequest struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"password" db:"password"`
	Address   string    `json:"address" db:"address"`
	Latitude  float64   `json:"late" db:"latitude"`
	Longitude float64   `json:"long" db:"longitude"`
}
type UserWithRoleRequest struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Role      string    `json:"role" db:"role"`
	Password  string    `json:"password" db:"password"`
	Address   string    `json:"address" db:"address"`
	Latitude  float64   `json:"late" db:"latitude"`
	Longitude float64   `json:"long" db:"longitude"`
}
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleSubAdmin Role = "sub_admin"
	RoleUser     Role = "user"
)

type Dish struct {
	Name        string  `db:"name" json:"name"`
	Description string  `db:"description" json:"description"`
	Price       float64 `db:"price" json:"price"`
}

type Restaurant struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Location  string    `db:"location" json:"location"`
	CreatedBy uuid.UUID `db:"created_by" json:"created_by"`
	Latitude  float64   `db:"latitude" json:"latitude"`
	Longitude float64   `db:"longitude" json:"longitude"`
}

type LoginRequest struct {
	Email    string `json:"email" `
	Password string `json:"password" `
}
type LoginIDAndPassword struct {
	UserID   uuid.UUID `json:"id" db:"id"`
	Email    string    `json:"email" db:"email" `
	Password string    `json:"string" db:"password" `
}
type LoginIDAndPasswordByRole struct {
	UserID   uuid.UUID `json:"id" db:"id"`
	Role     string    `json:"role" db:"role"`
	Email    string    `json:"email" db:"email"`
	Password string    `json:"password" db:"password"`
}
type AddressRequest struct {
	UserID    uuid.UUID `json:"userID" db:"user_id"`
	Address   string    `json:"address" db:"address"`
	Latitude  float64   `json:"latitude" db:"latitude"`
	Longitude float64   `json:"longitude" db:"longitude"`
}

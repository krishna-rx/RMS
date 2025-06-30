package models

import "github.com/google/uuid"

type UserRequest struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name" validate:"required"`
	Email     string    `json:"email" db:"email" validate:"required"`
	Password  string    `json:"password" db:"password" validate:"required"`
	Address   string    `json:"address" db:"address" validate:"required"`
	Latitude  float64   `json:"late" db:"latitude" validate:"required"`
	Longitude float64   `json:"long" db:"longitude" validate:"required"`
}

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
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AddressRequest struct {
	UserID    uuid.UUID `json:"userID" db:"user_id"`
	Address   string    `json:"address" db:"address"`
	Latitude  float64   `json:"latitude" db:"latitude"`
	Longitude float64   `json:"longitude" db:"longitude"`
}

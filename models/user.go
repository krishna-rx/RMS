package models

import "github.com/google/uuid"

type User struct {
	ID        string  `json:"id" db:"id"`
	Name      string  `json:"name" db:"name"`
	Email     string  `json:"email" db:"email"`
	Password  string  `json:"password" db:"password"`
	Address   string  `json:"address" db:"address"`
	Latitude  float64 `json:"lang" db:"latitude"`
	Longitude float64 `json:"long" db:"longitude"`
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

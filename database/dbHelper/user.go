package dbHelper

import (
	"fmt"
	"github.com/google/uuid"
	"rms/database"
)

func IsUserExists(email string) (bool, error) {
	SQL := `SELECT COUNT(id) > 0 as is_exist FROM users WHERE email = TRIM($1);`
	var check bool
	chkErr := database.DB.Get(&check, SQL, email)
	return check, chkErr
}
func CreateUser(name, email, password, address string, latitude, longitude float64) error {
	fmt.Println(latitude, longitude)
	_, err := database.DB.Exec(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3)`, name, email, password)
	var userID uuid.UUID
	err = database.DB.QueryRow(`SELECT id FROM users WHERE email = $1`, email).Scan(&userID)
	var roleID uuid.UUID
	err = database.DB.QueryRow(`SELECT id FROM roles WHERE role = 'USER'`).Scan(&roleID)
	if err != nil {
		return fmt.Errorf("USER role not found: %w", err)
	}
	_, err = database.DB.Exec(`
        INSERT INTO user_roles (user_id, roles_id) 
        VALUES ($1, $2) 
        ON CONFLICT DO NOTHING`,
		userID, roleID,
	)
	// store the address here
	if err != nil {
		return fmt.Errorf("USER role already exists: %w", err)
	}
	query := `INSERT INTO addresses (user_id, address, latitude, longitude)
	VALUES ($1, $2, $3, $4)`
	_, err = database.DB.Exec(query, userID, address, latitude, longitude)
	return err
}

func InsertIntoRestaurant(userID uuid.UUID, name string, location string, latitude float64, longitude float64) (uuid.UUID, error) {
	SQL := `INSERT INTO restaurants (name, location, created_by, latitude, longitude)
        VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var restaurantID uuid.UUID
	err := database.DB.QueryRow(SQL, name, location, userID, latitude, longitude).Scan(&restaurantID)
	if err != nil {
		return uuid.Nil, err
	}
	return restaurantID, nil
}

func CreateDishByRestaurantID(name, description string, price float64, restaurantID uuid.UUID, userID uuid.UUID) error {
	_, err := database.DB.Exec(`
		INSERT INTO dishes (name, description, price, restaurant_id, created_by)
		VALUES ($1, $2, $3, $4, $5)
	`, name, description, price, restaurantID, userID)
	return err
}

func CreateSubadminBYAdmin(name, email string, password string) error {
	_, err := database.DB.Exec(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3)`, name, email, password)
	if err != nil {
		return err
	}
	var userID uuid.UUID
	err = database.DB.QueryRow(`SELECT id FROM users WHERE email = $1`, email).Scan(&userID)
	var roleID uuid.UUID
	err = database.DB.QueryRow(`SELECT id FROM roles WHERE role = 'SUBADMIN'`).Scan(&roleID)
	if err != nil {
		return fmt.Errorf("USER role not found: %w", err)
	}
	_, err = database.DB.Exec(`
        INSERT INTO user_roles (user_id, roles_id) 
        VALUES ($1, $2) 
        ON CONFLICT DO NOTHING`,
		userID, roleID,
	)
	return err
}

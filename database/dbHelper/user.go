package dbHelper

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"rms/database"
	"rms/models"
	"time"
)

func IsUserExists(email string) (bool, error) {
	SQL := `SELECT COUNT(id) > 0 as is_exist FROM users WHERE email = TRIM($1) AND archived_at IS NULL;`
	var check bool
	chkErr := database.DB.Get(&check, SQL, email)
	return check, chkErr
}
func CreateUser(user models.UserRequest) error {
	err := database.Tx(database.DB, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3)`, user.Name, user.Email, user.Password)
		return err
	})
	if err != nil {
		return err
	}
	var userSessionResponse models.UserSessionResponse
	err = database.DB.QueryRow(`SELECT id FROM users WHERE email = $1`, user.Email).Scan(&userSessionResponse.UserID)
	if err != nil {
		return err
	}
	err = database.DB.QueryRow(`SELECT id FROM roles WHERE role = 'USER'`).Scan(&userSessionResponse.RolesID)
	if err != nil {
		return fmt.Errorf("user role not found: %w", err)
	}
	err = database.Tx(database.DB, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`
        INSERT INTO user_roles (user_id, roles_id) 
        VALUES ($1, $2) 
        ON CONFLICT DO NOTHING`,
			userSessionResponse.UserID, userSessionResponse.RolesID,
		)
		return err
	})
	if err != nil {
		return err
	}
	err = database.Tx(database.DB, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO addresses (user_id,address,latitude,longitude) VALUES ($1, $2 , $3 , $4)`, userSessionResponse.UserID, user.Address, user.Latitude, user.Longitude)
		return err
	})
	return err
}

func CreateUserSession(userID uuid.UUID) (uuid.UUID, error) {
	SQL := `INSERT INTO user_session (user_id) VALUES ($1) returning id`
	var userSessionID uuid.UUID
	err := database.DB.Get(&userSessionID, SQL, userID)
	if err != nil {
		return uuid.Nil, err
	}
	return userSessionID, nil
}
func LogoutUserByArchiving(userSessionToken string) error {
	_, err := database.DB.Exec(` UPDATE user_session SET archived_at = $1 WHERE id = $2`, time.Now(), userSessionToken)
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

package dbHelper

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"net/http"
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
func CreateUser(user *models.UserRequest) error {
	err := database.Tx(database.DB, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO users (name, email, password) VALUES ($1, $2, $3)`, user.Name, user.Email, user.Password)
		return err
	})
	if err != nil {
		logrus.Errorf("error inserting values in user table : %v", err)
		return err
	}
	var userID uuid.UUID
	err = database.DB.Get(&userID, `SELECT id FROM users WHERE email = $1`, user.Email)
	if err != nil {
		logrus.Errorf("error getting the user_id : %v", err)
		return err
	}
	role := models.RoleUser
	err = database.Tx(database.DB, func(tx *sqlx.Tx) error {
		_, err = tx.Exec(`INSERT INTO user_roles (user_id, role) VALUES ($1, $2)`, userID, role)
		return err
	})

	if err != nil {
		logrus.Errorf("error inserting roles values in user table : %v", err)
		return err
	}
	err = database.Tx(database.DB, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO addresses (user_id,address,latitude,longitude) VALUES ($1, $2 , $3 , $4)`, userID, user.Address, user.Latitude, user.Longitude)
		return err
	})
	return err
}
func GetAllRoleByID(roles *[]string, userID string) error {
	err := database.DB.Select(roles, "SELECT role from user_roles WHERE user_id = $1", userID)
	if err != nil {
		logrus.Errorf("error getting roles from user table : %v", err)
		return err
	}
	return err
}

func CreateUserWithRole(userWithRole *models.UserWithRoleRequest, r *http.Request) error {
	ctx := r.Context()
	rawUserID := ctx.Value("user_id")
	userID := rawUserID.(string)
	err := database.Tx(database.DB, func(tx *sqlx.Tx) error {
		_, err := tx.Exec(`INSERT INTO users (name, email, password,created_by) VALUES ($1, $2, $3,$4)`, userWithRole.Name, userWithRole.Email, userWithRole.Password, userID)
		return err
	})
	if err != nil {
		logrus.Errorf("error inserting values in user table : %v", err)
		return err
	}
	var userWithRoleID uuid.UUID
	err = database.DB.Get(&userWithRoleID, `SELECT id FROM users WHERE email = $1`, userWithRole.Email)
	if err != nil {
		logrus.Errorf("error getting the user_id : %v", err)
		return err
	}
	err = database.Tx(database.DB, func(tx *sqlx.Tx) error {
		_, err = tx.Exec(`INSERT INTO user_roles (user_id, role) VALUES ($1, $2)`, userWithRoleID, userWithRole.Role)
		return err
	})
	if err != nil {
		logrus.Errorf("error inserting role values in roles table : %v", err)
		return err
	}
	err = database.Tx(database.DB, func(tx *sqlx.Tx) error {
		_, err = tx.Exec(`INSERT INTO addresses (user_id , address, latitude , longitude) VALUES ($1 , $2, $3 , $4)`, userWithRoleID, userWithRole.Address, userWithRole.Latitude, userWithRole.Longitude)
		return err
	})
	if err != nil {
		logrus.Errorf("error inserting address in the address table : %v", err)
		return err
	}
	return err
}
func LookupUserCredentials(Creds *models.LoginIDAndPassword, email string) error {
	err := database.DB.Get(&Creds.UserID, Creds.Password, "SELECT id ,password FROM users WHERE email  = $1", email)
	return err
}

func LookupUserCredentialsByRole(Creds *models.LoginIDAndPasswordByRole, email string, role string, hasRole *bool) error {
	err := database.DB.Get(Creds, "SELECT id ,password FROM users WHERE email  = $1", email)
	if err != nil {
		logrus.Errorf("error getting the user_id : %v", err)
		return err
	}

	err = database.DB.Get(hasRole, "SELECT EXISTS (SELECT 1 FROM user_roles WHERE user_id = $1 AND role = $2 )", Creds.UserID, role)
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

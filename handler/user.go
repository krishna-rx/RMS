package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/umahmood/haversine"
	"net/http"
	"rms/database"
	"rms/database/dbHelper"
	"rms/models"
	"rms/utils"
	"time"
)

func Register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	fmt.Println(user)
	fmt.Println(user.Latitude, user.Longitude)
	exist, existErr := dbHelper.IsUserExists(user.Email)
	if existErr != nil {
		logrus.Errorf("IsUserExists query error: %s", existErr.Error())
	}
	if exist {
		http.Error(w, "user already exists", http.StatusBadRequest)
		return
	}
	hashedPassword, err := utils.HashedPassword(user.Password)
	if err != nil {
		http.Error(w, "failed to generate hashedPassword", http.StatusBadRequest)
		return
	}

	err = dbHelper.CreateUser(user.Name, user.Email, hashedPassword, user.Address, user.Latitude, user.Longitude)
	if err != nil {
		http.Error(w, "failed to create user", http.StatusBadRequest)
		return
	}
	err = json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
	}{"User created successfully"})
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "invalid JSON response", http.StatusBadRequest)
		return
	}
	var hashedPassword, userID string
	err = database.DB.QueryRow("SELECT id ,password FROM users WHERE email = $1", creds.Email).Scan(&userID, &hashedPassword)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}
	if !utils.CheckPasswordHash(creds.Password, hashedPassword) {
		logrus.Errorf("invalid password %+v", err)
	}
	SQL := `INSERT INTO user_session (user_id) VALUES ($1) returning id`
	var sessionID string
	err = database.DB.Get(&sessionID, SQL, userID)
	if err != nil {
		http.Error(w, "session not inserted in database", http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(map[string]string{
		"SessionToken": sessionID,
	})
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
func Logout(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("Authorization")
	if sessionToken == "" {
		http.Error(w, "missing session token", http.StatusBadRequest)
		return
	}
	var archived sql.NullTime
	err := database.DB.QueryRow(` SELECT archived_at FROM user_session WHERE id = $1`, sessionToken).Scan(&archived)
	if err != nil {
		http.Error(w, "fession not found", http.StatusNotFound)
		return
	}
	_, err = database.DB.Exec(` UPDATE user_session SET archived_at = $1 WHERE id = $2`, time.Now(), sessionToken)
	if err != nil {
		http.Error(w, "failed to update session", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{
		"message": "logout successfully",
	})
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id,name,email FROM users`
	rows, err := database.DB.Query(query)
	if err != nil {
		http.Error(w, "failed to get users", http.StatusInternalServerError)
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			http.Error(w, "failed to get user after scanning", http.StatusInternalServerError)
		}
		users = append(users, user)
	}
	json.NewEncoder(w).Encode(users)
}

func GetAllRestaurant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	restaurantIDStr := vars["restaurant_id"]
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
		return
	}
	var restaurants []models.Restaurant
	query := `SELECT name, location, created_by, latitude , longitude FROM restaurants WHERE id = $1`
	rows, err := database.DB.Query(query, restaurantID)
	if err != nil {
		http.Error(w, "no restaurants were found", http.StatusNotFound)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var restaurant models.Restaurant
		err = rows.Scan(&restaurant.Name, &restaurant.Location, &restaurant.CreatedBy, &restaurant.Latitude, &restaurant.Longitude)
		if err != nil {
			http.Error(w, "error iterating over restaurants", http.StatusNotFound)
			return
		}
		restaurants = append(restaurants, restaurant)
	}
	json.NewEncoder(w).Encode(restaurants)
}

func GetAllDishes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	restaurantIDStr := vars["restaurant_id"]
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		http.Error(w, "invalid restaurant ID", http.StatusBadRequest)
		return
	}
	query := `SELECT name,description,price FROM dishes WHERE restaurant_id = $1 `
	rows, err := database.DB.Query(query, restaurantID)
	if err != nil {
		http.Error(w, "no dishes were found", http.StatusNotFound)
		return
	}
	defer rows.Close()
	var dishes []models.Dish
	for rows.Next() {
		var dish models.Dish
		err = rows.Scan(&dish.Name, &dish.Description, &dish.Price)
		if err != nil {
			http.Error(w, "error iterating over dishes ", http.StatusNotFound)
			return
		}
		dishes = append(dishes, dish)
	}
	json.NewEncoder(w).Encode(dishes)
}

func GetDistanceFromUser(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	restaurantIDStr := query.Get("restaurant_id")
	address_id := query.Get("address_id")

	// Parse UUIDs
	restaurantID, err := uuid.Parse(restaurantIDStr)
	if err != nil {
		http.Error(w, "Invalid restaurant_id", http.StatusBadRequest)
		return
	}

	address_ID, err := uuid.Parse(address_id)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}
	var Rcordinate struct {
		long float64
		lat  float64
	}
	var Ucordinate struct {
		long float64
		lat  float64
	}
	err = database.DB.QueryRow(`SELECT latitude, longitude FROM addresses WHERE id = $1`, address_ID).
		Scan(&Ucordinate.lat, &Ucordinate.long)
	if err != nil {
		http.Error(w, "no cordinates were found for user", http.StatusNotFound)
	}
	err = database.DB.QueryRow(`SELECT latitude, longitude FROM restaurants WHERE id = $1`, restaurantID).
		Scan(&Rcordinate.lat, &Rcordinate.long)
	if err != nil {
		http.Error(w, "no cordinates for restaurants user", http.StatusNotFound)
	}
	Rcr := haversine.Coord{Lat: Rcordinate.lat, Lon: Rcordinate.long}
	Ucr := haversine.Coord{Lat: Ucordinate.lat, Lon: Ucordinate.long}
	_, km := haversine.Distance(Rcr, Ucr)
	json.NewEncoder(w).Encode(map[string]float64{
		"Distance from restaurant to user's location": km,
	})
}

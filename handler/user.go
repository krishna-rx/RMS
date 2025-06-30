package handler

import (
	"database/sql"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/umahmood/haversine"
	"net/http"
	"rms/database"
	"rms/database/dbHelper"
	"rms/models"
	"rms/utils"
)

var JSON = jsoniter.ConfigCompatibleWithStandardLibrary

func Register(w http.ResponseWriter, r *http.Request) {
	var user models.UserRequest
	err := JSON.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	validate := validator.New()
	err = validate.Struct(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	exist, existErr := dbHelper.IsUserExists(user.Email)
	if existErr != nil {
		http.Error(w, existErr.Error(), http.StatusBadRequest)
	}
	if exist {
		http.Error(w, "user already exits", http.StatusBadRequest)
		return
	}
	hashedPassword, err := utils.HashedPassword(user.Password)
	if err != nil {
		http.Error(w, "failed to generate hashedPassword", http.StatusBadRequest)
		return
	}
	user.Password = hashedPassword
	err = dbHelper.CreateUser(user)
	if err != nil {
		http.Error(w, "failed to create user", http.StatusBadRequest)
		return
	}
	response := models.UserResponse{
		Message:   "User created successfully",
		Name:      user.Name,
		Email:     user.Email,
		Address:   user.Address,
		Latitude:  user.Latitude,
		Longitude: user.Longitude,
	}
	err = JSON.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	var login models.LoginRequest
	err := JSON.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		http.Error(w, "invalid JSON response", http.StatusBadRequest)
		return
	}
	validate := validator.New()
	err = validate.Struct(login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var hashedPassword string
	var userID uuid.UUID
	err = database.DB.QueryRow("SELECT id ,password FROM users WHERE email = $1", login.Email).Scan(&userID, &hashedPassword)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}
	if !utils.CheckPasswordHash(login.Password, hashedPassword) {
		logrus.Errorf("invalid password %+v", err)
	}
	var userSessionID uuid.UUID
	userSessionID, err = dbHelper.CreateUserSession(userID)

	userResponse := models.LoginResponse{
		Message:               "User logged in",
		Name:                  login.Email,
		UserSessionIDResponse: userSessionID,
	}
	err = JSON.NewEncoder(w).Encode(userResponse)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	userSessionToken := r.Header.Get("Authorization")
	if userSessionToken == "" {
		http.Error(w, "missing session token", http.StatusBadRequest)
		return
	}
	var archived sql.NullTime
	err := database.DB.QueryRow(` SELECT archived_at FROM user_session WHERE id = $1`, userSessionToken).Scan(&archived)
	if err != nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	err = dbHelper.LogoutUserByArchiving(userSessionToken)

	err = JSON.NewEncoder(w).Encode(map[string]string{
		"message": "logout successfully",
	})
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id,name,email FROM users`
	rows, err := database.DB.Query(query)
	if err != nil {
		http.Error(w, "failed to get users", http.StatusInternalServerError)
	}
	defer rows.Close()
	var users []models.UserRequest
	for rows.Next() {
		var user models.UserRequest
		err = rows.Scan(&user.Name, &user.Email)
		if err != nil {
			http.Error(w, "failed to get user after scanning", http.StatusInternalServerError)
		}
		users = append(users, user)
	}
	JSON.NewEncoder(w).Encode(users)
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
	JSON.NewEncoder(w).Encode(restaurants)
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
	JSON.NewEncoder(w).Encode(dishes)
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
	JSON.NewEncoder(w).Encode(map[string]float64{
		"Distance from restaurant to user's location": km,
	})
}

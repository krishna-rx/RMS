package handler

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"rms/database"
	"rms/database/dbHelper"
	"rms/models"
	"rms/utils"
)

func CreateRestaurant(w http.ResponseWriter, r *http.Request) {
	roles := r.Context().Value("roles").([]string)
	if utils.HasRole(roles, "ADMIN") || utils.HasRole(roles, "SUBADMIN") {
		userIDVal := r.Context().Value("userID")
		userID, ok := userIDVal.(uuid.UUID)
		if !ok {
			http.Error(w, "userID missing or invalid", http.StatusUnauthorized)
			return
		}

		var restaurant models.Restaurant
		err := json.NewDecoder(r.Body).Decode(&restaurant)
		if err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
		}
		restaurantID, saveErr := dbHelper.InsertIntoRestaurant(userID, restaurant.Name, restaurant.Location, restaurant.Latitude, restaurant.Longitude)
		if saveErr != nil {
			logrus.Errorf("Error inserting restaurant to database%+v", saveErr)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Succeed":      "created restaurant",
			"restaurantID": restaurantID,
		})
	} else {
		http.Error(w, "not authorized to create resturant", http.StatusUnauthorized)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Not authorized to create restaurant",
		})
		return
	}
}
func CreateDishes(w http.ResponseWriter, r *http.Request) {
	roles := r.Context().Value("roles").([]string)
	if utils.HasRole(roles, "ADMIN") || utils.HasRole(roles, "SUBADMIN") {

		userIDVal := r.Context().Value("userID")
		userID, ok := userIDVal.(uuid.UUID)
		if !ok {
			http.Error(w, "userID missing or invalid", http.StatusUnauthorized)
			return
		}
		var Dishes models.Dish
		err := json.NewDecoder(r.Body).Decode(&Dishes)
		if err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
		}
		vars := mux.Vars(r)
		restaurantIDStr := vars["restaurant_id"]
		restaurantID, err := uuid.Parse(restaurantIDStr)
		if err != nil {
			http.Error(w, "Invalid restaurant ID", http.StatusBadRequest)
			return
		}

		fmt.Println("restaurantID:", restaurantID)
		fmt.Println("Dishes:", Dishes)
		fmt.Println("UserID:", userID)
		saveErr := dbHelper.CreateDishByRestaurantID(Dishes.Name, Dishes.Description, Dishes.Price, restaurantID, userID)
		if saveErr != nil {
			logrus.Errorf("error inserting restaurant to database%+v", saveErr)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Succeed": "created dish successfully",
		})

	} else {
		http.Error(w, "not authorized to create restaurant", http.StatusUnauthorized)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Not authorized to create restaurant",
		})
		return
	}
}
func CreateSubadmin(w http.ResponseWriter, r *http.Request) {
	roles := r.Context().Value("roles").([]string)
	if utils.HasRole(roles, "ADMIN") {
		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
		}
		hashedPassword, err := utils.HashedPassword(user.Password)
		if err != nil {
			http.Error(w, "failed to generate hashedPassword", http.StatusBadRequest)
			return
		}
		saveErr := dbHelper.CreateSubadminBYAdmin(user.Name, user.Email, hashedPassword)
		if saveErr != nil {
			logrus.Errorf("error inserting subadmin to database%+v", saveErr)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Succeed": "created subadmin successfully",
		})
	} else {
		http.Error(w, "not authorized to create restaurant", http.StatusUnauthorized)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Not authorized to create Subadmin",
		})
		return
	}
}

func GetAllSubadmin(w http.ResponseWriter, r *http.Request) {
	query := `
    SELECT name , email 
    FROM users AS u
    JOIN user_roles As ur ON ur.user_id = u.id
    JOIN roles AS r ON ur.roles_id = r.id
    WHERE r.role = 'SUBADMIN'`
	rows, err := database.DB.Query(query)
	if err != nil {
		http.Error(w, "error getting all subadmins", http.StatusBadRequest)
	}
	defer rows.Close()
	var users []models.User
	if rows == nil {
		http.Error(w, "no subadmins found", http.StatusBadRequest)
		return
	}
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.Name, &user.Email)
		if err != nil {
			http.Error(w, "error getting iterating all subadmins", http.StatusBadRequest)
		}
		users = append(users, user)
	}
	json.NewEncoder(w).Encode(users)
}

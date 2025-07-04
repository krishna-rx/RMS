package handler

import (
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/umahmood/haversine"
	"net/http"
	"rms/database"
	"rms/database/dbHelper"
	"rms/models"
	"rms/utils"
	"strings"
)

var json = utils.JSON

func UserRegister(w http.ResponseWriter, r *http.Request) {
	var user models.UserRequest
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "failed to parse request body")
		return
	}
	//for leading spaces
	user.Name = strings.TrimSpace(user.Name)
	user.Email = strings.TrimSpace(user.Email)
	if user.Name == "" || user.Email == "" || user.Password == "" || user.Address == "" || user.Latitude < 0 || user.Longitude < 0 {
		utils.ResponseError(w, http.StatusBadRequest, "name , email , password, role , address , latitude , longitude are mandatory ")
		return
	}
	exist, existErr := dbHelper.IsUserExists(user.Email)
	if existErr != nil {
		utils.ResponseError(w, http.StatusBadRequest, "error in checking existence of user")
		return
	}
	if exist {
		//error 409 status conflict
		utils.ResponseError(w, http.StatusConflict, "user with this email already exists")
		return
	}
	hashedPassword, err := utils.HashedPassword(user.Password)
	if err != nil {
		//500 internal server error
		utils.ResponseError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	user.Password = hashedPassword
	err = dbHelper.CreateUser(&user)

	if err != nil {
		// 500 internal server error
		utils.ResponseError(w, http.StatusInternalServerError, "failed to create user")
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
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "failed to encode response")
		return
	}
}

//func RegisterUser(w http.ResponseWriter, r *http.Request) {
//	var adminRequest models.RegisterAdminRequest
//	err := json.NewDecoder(r.Body).Decode(&adminRequest)
//	if err != nil {
//		utils.ResponseError(w, http.StatusBadRequest, "failed to parse request body")
//		return
//	}
//	if adminRequest.Name == "" || adminRequest.Email == "" || adminRequest.Role == "" || adminRequest.Password == "" || adminRequest.Address == "" || adminRequest.Latitude < 0 || adminRequest.Longitude < 0 {
//		utils.ResponseError(w, http.StatusBadRequest, "name , email , password, role , address , latitude , longitude are mandatory ")
//		return
//	}
//	exist, existErr := dbHelper.IsUserExists(adminRequest.Email)
//	if existErr != nil {
//		utils.ResponseError(w, http.StatusBadRequest, "error in checking existence of user")
//		return
//	}
//	if exist {
//		utils.ResponseError(w, http.StatusBadRequest, "user with this email already exists")
//		return
//	}
//	adminRequest.Password, err = utils.HashedPassword(adminRequest.Password)
//	if err != nil {
//		utils.ResponseError(w, http.StatusBadRequest, "failed to hash password")
//		return
//	}
//	//err := CreateUserWithRole(&adminRequest)
//
//}

// Login Normal User
func UserLogin(w http.ResponseWriter, r *http.Request) {
	var login models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "failed to parse request body for login")
		return
	}
	if login.Email == "" || login.Password == "" {
		utils.ResponseError(w, http.StatusBadRequest, "email and password are mandatory")
		return
	}
	var LoginDetails models.LoginIDAndPassword
	err = dbHelper.LookupUserCredentials(&LoginDetails, login.Email)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "failed to lookup user")
		return
	}
	if !utils.CheckPasswordHash(login.Password, LoginDetails.Password) {
		logrus.Errorf("invalid password %+v", err)
	}
	var userSessionID uuid.UUID
	userSessionID, err = dbHelper.CreateUserSession(LoginDetails.UserID)

	userResponse := models.LoginResponse{
		Message:               "User logged in",
		Name:                  login.Email,
		UserSessionIDResponse: userSessionID,
	}
	err = json.NewEncoder(w).Encode(userResponse)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// Login Admin
func LoginAsAdmin(w http.ResponseWriter, r *http.Request) {
	loginWithRole(w, r, "admin")
}
func loginWithRole(w http.ResponseWriter, r *http.Request, role string) {
	var login models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "failed to parse request body")
		return
	}
	if login.Email == "" || login.Password == "" {
		utils.ResponseError(w, http.StatusBadRequest, "email and password are mandatory")
		return
	}
	var loginDetailsByRole models.LoginIDAndPasswordByRole
	var hasRole bool
	err = dbHelper.LookupUserCredentialsByRole(&loginDetailsByRole, login.Email, role, &hasRole)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "failed to lookup user")
		return
	}
	if hasRole == false {
		utils.ResponseError(w, http.StatusBadRequest, "unauthorized user")
		return
	}
	tokenString, err := utils.GenerateJWT(&loginDetailsByRole)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "failed to generate JWT token")
		return
	}
	err = json.NewEncoder(w).Encode(map[string]string{
		"jwtToken": tokenString,
	})
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "failed to encode response")
	}
}

func RegisterUserWithRole(w http.ResponseWriter, r *http.Request) {
	var registerWithRole models.UserWithRoleRequest
	err := json.NewDecoder(r.Body).Decode(&registerWithRole)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "failed to parse request body")
		return
	}
	if registerWithRole.Name == "" || registerWithRole.Email == "" || registerWithRole.Password == "" || registerWithRole.Role == "" || registerWithRole.Address == "" || registerWithRole.Latitude <= 0 || registerWithRole.Longitude <= 0 {
		utils.ResponseError(w, http.StatusBadRequest, "name ,email,password,role , address ,latitude,longitude are mandatory fields ")
	}
	registerWithRole.Name = strings.TrimSpace(registerWithRole.Name)
	registerWithRole.Email = strings.TrimSpace(registerWithRole.Email)
	ctx := r.Context()
	rawRoles := ctx.Value("roles")
	roles, ok := rawRoles.([]models.Role)
	if !ok {
		utils.ResponseError(w, http.StatusBadRequest, "failed to get roles from the context")
	}
	if utils.CheckRoles(roles, registerWithRole.Role) == false {
		utils.ResponseError(w, http.StatusBadRequest, "dont have permission to create the user!")
		return
	}
	exist, existErr := dbHelper.IsUserExists(registerWithRole.Email)
	if existErr != nil {
		utils.ResponseError(w, http.StatusBadRequest, "error in checking existence of user")
		return
	}
	if exist {
		utils.ResponseError(w, http.StatusBadRequest, "user already exists")
		return
	}
	registerWithRole.Password, err = utils.HashedPassword(registerWithRole.Password)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "failed to hash password")
	}
	err = dbHelper.CreateUserWithRole(&registerWithRole, r)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "failed to create user with role")
		return
	}
	userWithRole := models.UserWithRoleResponse{
		Name:      registerWithRole.Name,
		Role:      registerWithRole.Role,
		Email:     registerWithRole.Email,
		Address:   registerWithRole.Address,
		Latitude:  registerWithRole.Latitude,
		Longitude: registerWithRole.Longitude,
	}
	err = json.NewEncoder(w).Encode(userWithRole)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "failed to encode response")
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

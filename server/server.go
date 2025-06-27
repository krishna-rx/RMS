package server

import (
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"rms/handler"
	"rms/middlewares"
)

type Server struct {
	Router *mux.Router
	Addr   string
}

func SetupRoutes(addr string) *Server {
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World in this world"))
	}).Methods("GET")

	r.HandleFunc("/register", handler.Register).Methods("POST")
	r.HandleFunc("/login", handler.Login).Methods("POST")
	r.HandleFunc("/restaurant/{restaurant_id}", handler.GetAllRestaurant).Methods("GET")
	r.HandleFunc("/restaurant/{restaurant_id}/getAllDishes", handler.GetAllDishes).Methods("GET")
	protected := r.PathPrefix("/").Subrouter()
	protected.Use(middlewares.AuthMiddleware, middlewares.CheckOwnership)
	protected.HandleFunc("/users", handler.GetAllUsers).Methods("GET")
	protected.HandleFunc("/createSubadmin", handler.CreateSubadmin).Methods("POST")
	protected.HandleFunc("/logout", handler.Logout).Methods("POST")
	protected.HandleFunc("/createRestaurant", handler.CreateRestaurant).Methods("POST")
	protected.HandleFunc("/restaurant/{restaurant_id}/createDishes", handler.CreateDishes).Methods("POST")
	protected.HandleFunc("/getAllSubadmins", handler.GetAllSubadmin).Methods("GET")
	protected.HandleFunc("/getDistance", handler.GetDistanceFromUser).Methods("GET")

	return &Server{
		Router: r,
		Addr:   addr,
	}
}
func (s *Server) Start() {
	err := http.ListenAndServe(s.Addr, s.Router)
	if err != nil {
		logrus.Errorf("Error starting server: %s", err)
	} else {
		logrus.Info("Server started at:", s.Addr)
	}
}

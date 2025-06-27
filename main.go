package main

import (
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"os"
	"rms/database"
	"rms/server"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		logrus.Error("Error loading .env file")
	}
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")
	consStr := "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbname + "?sslmode=disable"
	database.InitDBAndMirate(consStr)
	defer database.CloseDB()
	s := server.SetupRoutes(":8080")
	s.Start()
}

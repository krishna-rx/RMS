package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"os"
	"rms/database"
	"rms/server"
	"strings"
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

	var builder strings.Builder
	_, err = fmt.Fprintf(&builder, "postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
	if err != nil {
		return
	}
	database.InitDBAndMirate(builder.String())
	defer database.CloseDB()
	s := server.SetupRoutes(":8080")
	s.Start()
}

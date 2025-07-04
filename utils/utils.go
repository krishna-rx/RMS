package utils

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"rms/models"
	"time"
)

var JSON = jsoniter.ConfigCompatibleWithStandardLibrary

func HashedPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		fmt.Println("Compare error:", err)
	}
	return err == nil
}
func HasRole(roles []string, target string) bool {
	for _, r := range roles {
		if r == target {
			return true
		}
	}
	return false
}

type clientError struct {
	StatusCode    int    `json:"statusCode"`
	MessageToUser string `json:"messageToUser"`
}

func ResponseError(w http.ResponseWriter, statusCode int, messageToUser string) {
	logrus.Errorf("status : %d, message : %s", statusCode, messageToUser)
	clientErr := &clientError{
		StatusCode:    statusCode,
		MessageToUser: messageToUser,
	}
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(clientErr); err != nil {
		logrus.Errorf("failed to send the error %+v", err)
	}
}
func GenerateJWT(Creds *models.LoginIDAndPasswordByRole) (string, error) {

	claims := jwt.MapClaims{
		"user_id": Creds.UserID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
		"role":    Creds.Role,
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := []byte(os.Getenv("JWT_SECRET"))
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func CheckRoles(roles []models.Role, target string) bool {
	for _, r := range roles {
		switch string(r) {
		case "admin":
			{
				if target == "admin" || target == "sub_admin" || target == "user" {
					return true
				}
			}
		case "sub_admin":
			{
				if target == "sub_admin" || target == "user" {
					return true
				}
			}
		}
	}
	return false
}

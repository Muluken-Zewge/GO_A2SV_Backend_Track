package infrastructure

import (
	"errors"
	"fmt"
	"os"
	domain "taskmanager/Domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userId string, userName string, role domain.UserRole) (string, error) {

	// get jwt secret from env varaiable
	jwtSecret := os.Getenv("JWT_SECRET")

	// 2. VALIDATION: Check if the secret is missing or empty
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET environment variable is not set or empty")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   userId,
		"user_name": userName,
		"role":      role,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	})

	// sign the token with the secret key
	jwtToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}

	return jwtToken, nil
}

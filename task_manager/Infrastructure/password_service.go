package infrastructure

import (
	"fmt"
	domain "taskmanager/Domain"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {

	// hash the password with bcrypt package
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedPassword), nil
}

func ComparePassword(storedPassword string, insertedPassword string) error {

	// compare user insereted password with the stored one
	err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(insertedPassword))
	if err != nil {
		return domain.ErrValidation
	}

	return nil
}

package infrastructure_test

import (
	"errors"
	domain "taskmanager/Domain"
	infrastructure "taskmanager/Infrastructure"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword_Success(t *testing.T) {
	// ARRANGE
	password := "securepassword"

	// ACT
	hashed, err := infrastructure.HashPassword(password)

	// ASSERT: 1. check the result of the function
	require.NoError(t, err, "HashPassword shouldn't return an error")
	assert.NotEqual(t, hashed, password, "the password and it's hash shouldn't be equal")

	// ASSERT: 2. Verify the hash is a valid bcrypt hash by comparing it back to the original password
	err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	assert.NoError(t, err, "the generated hash should be valid for comparison")
}

func TestComparePassword_Match(t *testing.T) {

	// ARRANGE: Hash a password to simulate a stored hash
	plainPassword := "password"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)

	// ACT & ASSERT: Compare the plain password against the stored hash.
	err := infrastructure.ComparePassword(string(hashedPassword), plainPassword)

	assert.NoError(t, err, "ComparePassword should return nil for a matching password")
}

func TestComparePassword_Mismatch(t *testing.T) {

	// ARRANGE: Hash a correct password
	correctPassword := "correctpassword"
	wrongPassword := "incorrectpassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

	// ACT & ASSERT: Compare a different password against the stored hash.
	err := infrastructure.ComparePassword(string(hashedPassword), wrongPassword)

	assert.Error(t, err, "ComparePassword should return an error for a mismatch")
	// CRITICAL ASSERT: Check that the error is the expected domain error (ErrValidation).
	assert.True(t, errors.Is(err, domain.ErrValidation), "Mismatch error should be domain.ErrValidation")
}

package infrastructure_test

import (
	"os"
	domain "taskmanager/Domain"
	infrastructure "taskmanager/Infrastructure"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// temporarily sets the JWT_SECRET for testing
func setTestSecret(secret string) func() {
	// Set the required environment variable
	os.Setenv("JWT_SECRET", secret)

	// Return a cleanup function
	return func() {
		os.Unsetenv("JWT_SECRET")
	}
}

func TestGenerateJWT_Success(t *testing.T) {

	// ARRANGE 1: Set a secret and ensure it is cleaned up.
	testSecret := "test_secret_key_123"
	cleanup := setTestSecret(testSecret)
	defer cleanup()

	// ARRANGE 2: Define expected claims data.
	expectedUserID := "a1b2c3d4-e5f6-7890-1234-567890abcdef"
	expectedUserName := "testuser"
	expectedRole := domain.RoleAdmin

	// ACT: Generate the JWT token.
	tokenString, err := infrastructure.GenerateJWT(expectedUserID, expectedUserName, expectedRole)

	// ASSERT 1: Check for no error and that a token string was returned
	require.NoError(t, err, "GenerateJWT should not return an error")
	assert.NotEmpty(t, tokenString, "Generated token string should not be empty")

	// ASSERT 2: Parse the token to verify claims
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		assert.Equal(t, jwt.SigningMethodHS256, token.Method)
		return []byte(testSecret), nil
	})

	// ASSERT 3: Check parsing success and claims content
	assert.NoError(t, err, "Token parsing should succeed")
	assert.True(t, token.Valid, "Token should be valid")

	claims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok, "Claims should be of type jwt.MapClaims")

	// Verify the custom claims
	assert.Equal(t, expectedUserID, claims["user_id"], "Claim 'user_id' mismatch")
	assert.Equal(t, expectedUserName, claims["user_name"], "Claim 'user_name' mismatch")
	assert.Equal(t, expectedRole, domain.UserRole(claims["role"].(float64)), "Claim 'role' mismatch")
}

func TestGenerateJWT_NoSecret(t *testing.T) {
	// ARRANGE: Ensure JWT_SECRET is NOT set (Unset any potential residual value).
	os.Unsetenv("JWT_SECRET")

	// ACT: Generate the JWT token.
	tokenString, err := infrastructure.GenerateJWT("id", "user", domain.RoleUser)

	// ASSERT: The JWT library will likely panic/error when signing with an empty key,
	// We verify an error is returned and no token is present.
	assert.Error(t, err, "GenerateJWT should return an error when JWT_SECRET is missing or empty")
	assert.Empty(t, tokenString, "Token string should be empty on error")
}

package infrastructure_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domain "taskmanager/Domain"
	infrastructure "taskmanager/Infrastructure" // The package being tested

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// --- Helper Functions ---

const testSecret = "supersecretkeyforauth"
const testUserID = "01234567-89ab-cdef-0123-456789abcdef"

// generateTestToken creates a valid, signed JWT for testing
func generateTestToken(t *testing.T, userID string, role domain.UserRole, expiration time.Duration) string {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    float64(role), // JWT claims usually use float64 for numbers
		"exp":     time.Now().Add(expiration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSecret))
	assert.NoError(t, err, "Failed to sign test token")
	return tokenString
}

// executeMiddleware executes the middleware function in a test context
func executeMiddleware(handler gin.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new recording response writer
	w := httptest.NewRecorder()

	// Create a Gin context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Use a fake next handler to check if the middleware called it
	nextHandler := gin.HandlerFunc(func(c *gin.Context) {
		c.Status(http.StatusOK) // Mark success if we reach here
	})

	// Execute the middleware chain
	handler(c)
	nextHandler(c)

	// Add the check to the recorder's result
	if !c.IsAborted() {
		w.Header().Set("X-Next-Called", "true")
	} else {
		w.Header().Set("X-Next-Called", "false")
	}

	return w
}

// --- 1. Testing AuthMiddleware ---

func TestAuthMiddleware_Success(t *testing.T) {
	// ARRANGE: Create a valid, unexpired token for a regular user
	validToken := generateTestToken(t, testUserID, domain.RoleUser, time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)

	// ACT
	middleware := infrastructure.AuthMiddleware(testSecret)
	w := executeMiddleware(middleware, req)

	// ASSERT
	assert.Equal(t, http.StatusOK, w.Code, "Should pass authentication and call next()")
	assert.Equal(t, "true", w.Header().Get("X-Next-Called"), "Middleware should not abort")
}

func TestAuthMiddleware_Fail_NoHeader(t *testing.T) {
	// ARRANGE: No Authorization header
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// ACT
	middleware := infrastructure.AuthMiddleware(testSecret)
	w := executeMiddleware(middleware, req)

	// ASSERT
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "false", w.Header().Get("X-Next-Called"), "Middleware must abort")

	var responseBody map[string]string
	json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.Equal(t, "Authorization header is required", responseBody["error"])
}

func TestAuthMiddleware_Fail_InvalidFormat(t *testing.T) {
	// ARRANGE: Header present but wrong format (e.g., missing 'Bearer')
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token somevalue") // Should be "Bearer"

	// ACT
	middleware := infrastructure.AuthMiddleware(testSecret)
	w := executeMiddleware(middleware, req)

	// ASSERT
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "false", w.Header().Get("X-Next-Called"), "Middleware must abort")

	var responseBody map[string]string
	json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.Equal(t, "Invalid authorization header", responseBody["error"])
}

func TestAuthMiddleware_Fail_ExpiredToken(t *testing.T) {
	// ARRANGE: Create an expired token
	expiredToken := generateTestToken(t, testUserID, domain.RoleUser, -time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)

	// ACT
	middleware := infrastructure.AuthMiddleware(testSecret)
	w := executeMiddleware(middleware, req)

	// ASSERT
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var responseBody map[string]string
	json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.Equal(t, "Invalid JWT", responseBody["error"])
}

func TestAuthMiddleware_Fail_WrongSecret(t *testing.T) {
	// ARRANGE: Token signed with the correct secret, but middleware uses a wrong secret
	validToken := generateTestToken(t, testUserID, domain.RoleUser, time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)

	// ACT
	middleware := infrastructure.AuthMiddleware("wrong-secret")
	w := executeMiddleware(middleware, req)

	// ASSERT
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "false", w.Header().Get("X-Next-Called"))

	var responseBody map[string]string
	json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.Equal(t, "Invalid JWT", responseBody["error"])
}

// --- 2. Testing AuthorizationMiddleware ---

// The Authorization middleware relies on 'role' being set in the context by AuthMiddleware.
// We must manually create the Gin context and set the 'role' value.
func executeAuthorizationMiddleware(requiredRole domain.UserRole, contextRole domain.UserRole) *httptest.ResponseRecorder {
	// 1. Arrange Context and Request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	// 2. Simulate AuthMiddleware having run successfully
	c.Set("role", contextRole)

	// 3. Act
	middleware := infrastructure.AuthorizationMiddleware(requiredRole)

	// Create a fake next handler
	nextHandler := gin.HandlerFunc(func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Execute the middleware chain
	middleware(c)
	nextHandler(c)

	if !c.IsAborted() {
		w.Header().Set("X-Next-Called", "true")
	} else {
		w.Header().Set("X-Next-Called", "false")
	}
	return w
}

func TestAuthorizationMiddleware_Success_AdminAccess(t *testing.T) {
	// ARRANGE: Admin (RoleAdmin) accessing an Admin-required route (RoleAdmin)
	w := executeAuthorizationMiddleware(domain.RoleAdmin, domain.RoleAdmin)

	// ASSERT
	assert.Equal(t, http.StatusOK, w.Code, "Admin should access Admin route")
	assert.Equal(t, "true", w.Header().Get("X-Next-Called"))
}

func TestAuthorizationMiddleware_Success_AdminAccessUserRoute(t *testing.T) {
	// ARRANGE: Admin (RoleAdmin) accessing a User-required route (RoleUser)
	w := executeAuthorizationMiddleware(domain.RoleUser, domain.RoleAdmin)

	// ASSERT (Admin > User, so access granted)
	assert.Equal(t, http.StatusOK, w.Code, "Admin should access User route (higher privilege)")
	assert.Equal(t, "true", w.Header().Get("X-Next-Called"))
}

func TestAuthorizationMiddleware_Fail_UserAccessAdminRoute(t *testing.T) {
	// ARRANGE: User (RoleUser) accessing an Admin-required route (RoleAdmin)
	w := executeAuthorizationMiddleware(domain.RoleAdmin, domain.RoleUser)

	// ASSERT (User < Admin, so access denied)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, "false", w.Header().Get("X-Next-Called"))

	var responseBody map[string]string
	json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.Equal(t, "Insufficient permissions", responseBody["error"])
}

func TestAuthorizationMiddleware_Fail_RoleMissing(t *testing.T) {
	// ARRANGE: Context is missing the 'role' key (simulating AuthMiddleware failure/not running)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	// ACT
	middleware := infrastructure.AuthorizationMiddleware(domain.RoleAdmin)
	middleware(c)

	// ASSERT
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var responseBody map[string]string
	json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.Equal(t, "Authentication required", responseBody["error"])
}

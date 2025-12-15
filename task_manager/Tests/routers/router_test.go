package router_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"taskmanager/Delivery/router"
	domain "taskmanager/Domain"
	"taskmanager/Tests/mocks"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Environment and Token Constants ---

const testSecret = "test_router_secret_key"
const adminUserID = "admin-1234-abcd"
const standardUserID = "user-5678-efgh"

// --- Setup and Helper Functions ---

func SetupTestRouter(t *testing.T) (*gin.Engine, *mocks.MockTaskUsecase, *mocks.MockUserUsecase) {
	// Must set JWT_SECRET for middleware to initialize correctly
	os.Setenv("JWT_SECRET", testSecret)

	// Create mocks
	// Note: We are using mock interfaces here, but the implementation is identical to the mock setup in the previous response.
	taskUsecaseMock := new(mocks.MockTaskUsecase)
	userUsecaseMock := new(mocks.MockUserUsecase)

	// Create router
	r := router.SetupRouter(taskUsecaseMock, userUsecaseMock)

	// Ensure cleanup
	t.Cleanup(func() { os.Unsetenv("JWT_SECRET") })

	return r, taskUsecaseMock, userUsecaseMock
}

// generateTestToken creates a valid, signed JWT for testing
func generateTestToken(t *testing.T, userID string, role domain.UserRole) string {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    float64(role), // JWT claims use float64 for numbers
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSecret))
	assert.NoError(t, err, "Failed to sign test token")
	return tokenString
}

// makeRequest is a utility to execute a request
func makeRequest(r *gin.Engine, method, url, token string, body ...interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Reader
	if len(body) > 0 && body[0] != nil {
		b, err := json.Marshal(body[0])
		if err != nil {
			reqBody = bytes.NewReader(nil)
		} else {
			reqBody = bytes.NewReader(b)
		}
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, url, reqBody)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if len(body) > 0 && body[0] != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// --- Router and Middleware Tests ---

func TestRouter_TaskReadRoutes_RequireAuth(t *testing.T) {
	r, taskMock, _ := SetupTestRouter(t)

	// Case 1: GET /api/v1/tasks - No Token (Should fail AuthMiddleware)
	w := makeRequest(r, http.MethodGet, "/api/v1/tasks", "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	taskMock.AssertNotCalled(t, "RetrieveAllTasks", mock.Anything)

	// Case 2: GET /api/v1/tasks/:id - Valid User Token (Should Pass)
	token := generateTestToken(t, standardUserID, domain.RoleUser)
	taskMock.EXPECT().RetrieveTaskByID(mock.Anything, "1").Return(domain.Task{}, nil)
	w = makeRequest(r, http.MethodGet, "/api/v1/tasks/1", token)
	assert.Equal(t, http.StatusOK, w.Code)
	taskMock.AssertExpectations(t)
}

func TestRouter_TaskWriteRoutes_RequireAdmin(t *testing.T) {
	r, taskMock, _ := SetupTestRouter(t)

	// 1. Attempt POST with Regular User Token (Should fail AuthorizationMiddleware)
	userToken := generateTestToken(t, standardUserID, domain.RoleUser)
	w := makeRequest(r, http.MethodPost, "/api/v1/tasks", userToken)

	assert.Equal(t, http.StatusForbidden, w.Code, "User should be forbidden from POST /tasks")
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Insufficient permissions", response["error"])
	taskMock.AssertNotCalled(t, "CreateTask", mock.Anything, mock.Anything)

	// 2. Attempt DELETE with Admin User Token (Should Pass)
	adminToken := generateTestToken(t, adminUserID, domain.RoleAdmin)
	taskMock.EXPECT().RemoveTask(mock.Anything, "1").Return(nil)

	w = makeRequest(r, http.MethodDelete, "/api/v1/tasks/1", adminToken)
	assert.Equal(t, http.StatusOK, w.Code, "Admin should be allowed to DELETE /tasks/:id")
	taskMock.AssertExpectations(t)
}

func TestRouter_UserPromoteRoute_RequireAdmin(t *testing.T) {
	r, _, userMock := SetupTestRouter(t)

	// 1. Attempt PATCH with Regular User Token (Should fail AuthorizationMiddleware)
	userToken := generateTestToken(t, standardUserID, domain.RoleUser)
	w := makeRequest(r, http.MethodPatch, "/api/v1/user/123/promote", userToken)

	assert.Equal(t, http.StatusForbidden, w.Code, "User should be forbidden from promoting others")
	userMock.AssertNotCalled(t, "PromoteUser", mock.Anything, mock.Anything)

	// 2. Attempt PATCH with Admin User Token (Should Pass)
	adminToken := generateTestToken(t, adminUserID, domain.RoleAdmin)
	userMock.EXPECT().PromoteUser(mock.Anything, "123").Return(domain.User{}, nil)

	w = makeRequest(r, http.MethodPatch, "/api/v1/user/123/promote", adminToken)
	assert.Equal(t, http.StatusOK, w.Code, "Admin should be allowed to PATCH /user/:id/promote")
	userMock.AssertExpectations(t)
}

func TestRouter_PublicRoutes_NoAuthRequired(t *testing.T) {
	r, _, userMock := SetupTestRouter(t)
	credentials := domain.Credentials{UserName: "test", Password: "p"}

	// Case 1: POST /api/v1/user/register
	userMock.EXPECT().RegisterUser(mock.Anything, mock.Anything, mock.Anything).Return(domain.User{ID: uuid.UUID{}, UserName: "test"}, nil)
	w := makeRequest(r, http.MethodPost, "/api/v1/user/register", "", credentials)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Case 2: POST /api/v1/user/login
	userMock.EXPECT().AuthenticateUser(mock.Anything, mock.Anything, mock.Anything).Return("token", nil)
	w = makeRequest(r, http.MethodPost, "/api/v1/user/login", "", credentials)
	assert.Equal(t, http.StatusOK, w.Code)

	userMock.AssertExpectations(t)
}

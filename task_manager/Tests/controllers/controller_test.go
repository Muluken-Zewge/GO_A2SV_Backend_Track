package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"taskmanager/Delivery/controllers"
	domain "taskmanager/Domain"
	"taskmanager/Tests/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Setup Test Context and Helpers ---

// setupTestContext creates a new Gin context and recorder for testing a single handler.
func setupTestContext(method, url string, body interface{}, params gin.Params) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var reqBody *bytes.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewReader(jsonBody)
	} else {
		reqBody = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = params

	// Shorten timeout for tests
	c.Request = c.Request.WithContext(context.Background())

	return c, w
}

// --- Task Controller Tests ---

func TestTaskController_GetTasks_Success(t *testing.T) {
	mockUsecase := new(mocks.MockTaskUsecase)
	controller := controllers.NewTaskController(mockUsecase)
	c, w := setupTestContext(http.MethodGet, "/tasks", nil, nil)

	expectedTasks := []domain.Task{{ID: "1", Title: "Test"}}
	mockUsecase.EXPECT().RetrieveAllTasks(mock.Anything).Return(expectedTasks, nil)

	controller.GetTasks(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string][]domain.Task
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Len(t, response["tasks"], 1)

	mockUsecase.AssertExpectations(t)
}

func TestTaskController_GetTaskById_Fail_NotFound(t *testing.T) {
	mockUsecase := new(mocks.MockTaskUsecase)
	controller := controllers.NewTaskController(mockUsecase)

	params := gin.Params{{Key: "id", Value: "999"}}
	c, w := setupTestContext(http.MethodGet, "/tasks/999", nil, params)

	mockUsecase.EXPECT().RetrieveTaskByID(mock.Anything, "999").Return(domain.Task{}, domain.ErrNotFound)

	controller.GetTaskById(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "task not found", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestTaskController_CreatTask_Fail_Validation(t *testing.T) {
	mockUsecase := new(mocks.MockTaskUsecase)
	controller := controllers.NewTaskController(mockUsecase)

	validTask := domain.Task{Title: "Title"}
	c, w := setupTestContext(http.MethodPost, "/tasks", validTask, nil)

	// Mock Usecase returning a validation error
	mockUsecase.EXPECT().
		CreateTask(mock.Anything, validTask).
		Return(domain.Task{}, fmt.Errorf("%w: title, description and status are required", domain.ErrValidation))

	controller.CreatTask(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "title, description and status are required")

	mockUsecase.AssertExpectations(t)
}

func TestTaskController_DeleteTask_Success(t *testing.T) {
	mockUsecase := new(mocks.MockTaskUsecase)
	controller := controllers.NewTaskController(mockUsecase)

	params := gin.Params{{Key: "id", Value: "123"}}
	c, w := setupTestContext(http.MethodDelete, "/tasks/123", nil, params)

	mockUsecase.EXPECT().RemoveTask(mock.Anything, "123").Return(nil)

	controller.DeleteTask(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Task deleted successfully", response["message"])

	mockUsecase.AssertExpectations(t)
}

// --- User Controller Tests ---

func TestUserController_RegisterUser_Fail_AlreadyExists(t *testing.T) {
	mockUsecase := new(mocks.MockUserUsecase)
	controller := controllers.NewUserController(mockUsecase)

	credentials := domain.Credentials{UserName: "taken", Password: "p"}
	c, w := setupTestContext(http.MethodPost, "/user/register", credentials, nil)

	// Mock Usecase returning a wrapped ErrAleadyExists
	mockUsecase.EXPECT().RegisterUser(mock.Anything, "taken", "p").Return(domain.User{}, domain.ErrAleadyExists)

	controller.RegisterUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, domain.ErrAleadyExists.Error(), response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestUserController_AuthenticateUser_Fail_InvalidCredentials(t *testing.T) {
	mockUsecase := new(mocks.MockUserUsecase)
	controller := controllers.NewUserController(mockUsecase)

	credentials := domain.Credentials{UserName: "baduser", Password: "badpassword"}
	c, w := setupTestContext(http.MethodPost, "/user/login", credentials, nil)

	// Mock Usecase failing due to validation (bad password) or not found (bad username)
	mockUsecase.EXPECT().AuthenticateUser(mock.Anything, "baduser", "badpassword").Return("", domain.ErrValidation)

	controller.AuthenticateUser(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid username or password", response["error"])

	mockUsecase.AssertExpectations(t)
}

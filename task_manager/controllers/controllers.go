package controllers

import (
	"net/http"
	"strings"
	"taskmanager/data"
	"taskmanager/models"

	"github.com/gin-gonic/gin"
)

// --- TASK CONTROLLER ---

type TaskController struct {
	Service *data.TaskService
}

// NewTaskController creates a new instance of the controller
func NewTaskController(tS *data.TaskService) *TaskController {
	return &TaskController{
		Service: tS,
	}
}

func (t *TaskController) GetTasks(c *gin.Context) {
	allTasks, err := t.Service.GetAllTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"tasks": allTasks})
}

func (t *TaskController) GetTaskById(c *gin.Context) {
	id := c.Param("id")
	task, err := t.Service.GetTaskById(id)
	if err != nil {
		errorMessage := err.Error()
		if errorMessage == "task not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": errorMessage})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": task})
}

func (t *TaskController) CreatTask(c *gin.Context) {
	var newTask models.Task
	if err := c.ShouldBindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// validation
	if newTask.Title == "" || newTask.Description == "" || newTask.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title, Description, and Status are required fields"})
		return
	}
	createdTask, err := t.Service.CreateTask(newTask)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Task created successfully", "Task": createdTask})
}

func (t *TaskController) UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var updatedTask models.Task
	if err := c.ShouldBindJSON(&updatedTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task, err := t.Service.UpdateTask(id, updatedTask)
	if err != nil {
		errorMessage := err.Error()
		if errorMessage == "task not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": errorMessage})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully", "Updated task": task})

}

func (t *TaskController) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	err := t.Service.DeleteTask(id)
	if err != nil {
		errorMessage := err.Error()
		if errorMessage == "task not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": errorMessage})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// --- USER CONTROLLER ---

type UserController struct {
	service *data.UserService
}

func NewUserController(us *data.UserService) *UserController {
	return &UserController{
		service: us,
	}
}

func (u *UserController) RegisterUser(c *gin.Context) {
	// read and bind request body to user variable
	var userCredential models.Credentials
	if err := c.ShouldBindJSON(&userCredential); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// call the appropriate service function
	registeredUser, err := u.service.RegisterUser(userCredential)
	if err != nil {
		errorMessage := err.Error()

		//Check for expected client-side validation errors
		if strings.Contains(errorMessage, "password should be at least") || strings.Contains(errorMessage, "User name already exist") {
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMessage})
			return
		}

		// Default to 500 for actual server/database errors
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user due to a server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created successfully", "user": registeredUser})
}

func (u *UserController) AuthenticateUser(c *gin.Context) {
	// read and bind request body to user variable
	var userCredential models.Credentials
	if err := c.ShouldBindJSON(&userCredential); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// call the appropriate service function
	token, err := u.service.AuthenticateUser(userCredential)
	if err != nil {
		errorMessage := err.Error()

		// Check for invalid credentials (User's fault)
		if strings.Contains(errorMessage, "invalid credential") {
			// Authentication failed -> 401 Unauthorized
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		// Default to 500 for actual server/database errors
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed due to a server issue"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "message": "Login successfully"})
}

func (u *UserController) PromoteUser(c *gin.Context) {
	id := c.Param("id")
	user, err := u.service.PromoteUser(id)
	if err != nil {
		errorMessage := err.Error()
		if strings.Contains(errorMessage, "user not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": errorMessage})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user status updated successfully", "updatedUser": user})

}

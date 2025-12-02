package controllers

import (
	"context"
	"errors"
	"net/http"
	domain "taskmanager/Domain"
	usecases "taskmanager/Usecases"
	"time"

	"github.com/gin-gonic/gin"
)

// --- TASK CONTROLLER ---

type TaskController struct {
	taskUsecase usecases.TaskUsecase
}

// NewTaskController creates a new instance of the controller
func NewTaskController(tu usecases.TaskUsecase) *TaskController {
	return &TaskController{
		taskUsecase: tu,
	}
}

func (t *TaskController) GetTasks(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	allTasks, err := t.taskUsecase.RetrieveAllTasks(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tasks": allTasks})
}

func (t *TaskController) GetTaskById(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id := c.Param("id")
	task, err := t.taskUsecase.RetrieveTaskByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": task})
}

func (t *TaskController) CreatTask(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var newTask domain.Task
	if err := c.ShouldBindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdTask, err := t.taskUsecase.CreateTask(ctx, newTask)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Task created successfully", "Task": createdTask})
}

func (t *TaskController) UpdateTask(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id := c.Param("id")
	var updatedTask domain.Task
	if err := c.ShouldBindJSON(&updatedTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task, err := t.taskUsecase.ModifyTask(ctx, id, updatedTask)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully", "Updated task": task})

}

func (t *TaskController) DeleteTask(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id := c.Param("id")
	err := t.taskUsecase.RemoveTask(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// --- USER CONTROLLER ---

type UserController struct {
	userUsecase usecases.UserUsecase
}

func NewUserController(uu usecases.UserUsecase) *UserController {
	return &UserController{
		userUsecase: uu,
	}
}

func (u *UserController) RegisterUser(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// read and bind request body to user variable
	var userCredential domain.Credentials
	if err := c.ShouldBindJSON(&userCredential); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// call the appropriate usecase method
	registeredUser, err := u.userUsecase.RegisterUser(ctx, userCredential.UserName, userCredential.Password)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, domain.ErrAleadyExists) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Default to 500 for actual server/database errors
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user due to a server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created successfully", "user": registeredUser})
}

func (u *UserController) AuthenticateUser(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// read and bind request body to user variable
	var userCredential domain.Credentials
	if err := c.ShouldBindJSON(&userCredential); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// call the appropriate service function
	token, err := u.userUsecase.AuthenticateUser(ctx, userCredential.UserName, userCredential.Password)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
			return
		}
		// Default to 500 for actual server/database errors
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed due to a server issue"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "message": "Login successfully"})
}

func (u *UserController) PromoteUser(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id := c.Param("id")
	user, err := u.userUsecase.PromoteUser(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user status updated successfully", "updatedUser": user})

}

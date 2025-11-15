package controllers

import (
	"net/http"
	"taskmanager/data"
	"taskmanager/models"

	"github.com/gin-gonic/gin"
)

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

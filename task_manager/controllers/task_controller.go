package controllers

import (
	"net/http"
	"taskmanager/data"
	"taskmanager/models"

	"github.com/gin-gonic/gin"
)

func GetTasks(c *gin.Context) {
	allTasks := data.GetAllTasks()
	c.JSON(http.StatusOK, gin.H{"tasks": allTasks})
}

func GetTaskById(c *gin.Context) {
	id := c.Param("id")
	task, err := data.GetTaskById(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": task})
}

func CreatTask(c *gin.Context) {
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
	createdTask := data.CreateTask(newTask)
	c.JSON(http.StatusCreated, gin.H{"message": "Task created successfully", "Task": createdTask})
}

func UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var updatedTask models.Task
	if err := c.ShouldBindJSON(&updatedTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task, err := data.UpdateTask(id, updatedTask)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully", "Updated task": task})

}

func DeleteTask(c *gin.Context) {
	id := c.Param("id")
	err := data.DeleteTask(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

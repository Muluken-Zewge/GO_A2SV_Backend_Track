package data

import (
	"net/http"
	"strconv"
	"time"

	"taskmanager/models"

	"github.com/gin-gonic/gin"
)

var tasks = []models.Task{}
var taskId = 1

func getTasks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func getTaskById(c *gin.Context) {
	id := c.Param("id")
	for _, task := range tasks {
		if id == task.ID {
			c.JSON(http.StatusOK, task)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
}

func updateTask(c *gin.Context) {
	id := c.Param("id")
	var updatedTask models.Task

	if err := c.ShouldBindJSON(&updatedTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i, task := range tasks {
		if id == task.ID {
			// update only specified feilds
			if updatedTask.Title != "" {
				tasks[i].Title = updatedTask.Title
			}
			if updatedTask.Description != "" {
				tasks[i].Description = updatedTask.Description
			}
			if !updatedTask.DueDate.IsZero() {
				tasks[i].DueDate = updatedTask.DueDate
			}
			if updatedTask.Status != "" {
				tasks[i].Status = updatedTask.Status
			}

			c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully", "Updated Task": task})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})

}

func deleteTask(c *gin.Context) {
	id := c.Param("id")
	for i, task := range tasks {
		if id == task.ID {
			tasks = append(tasks[:i], tasks[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
}

func createTask(c *gin.Context) {
	var newTask models.Task

	if err := c.ShouldBindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if newTask.Title == "" || newTask.Description == "" || newTask.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title, description, and status are required fields"})
		return
	}

	newTask.ID = strconv.Itoa(taskId)
	newTask.DueDate = time.Now()
	taskId++ // increment for next task

	tasks = append(tasks, newTask)

	c.JSON(http.StatusCreated, gin.H{"message": "Task created successfully", "task": newTask})

}

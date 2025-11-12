package data

import (
	"errors"
	"strconv"
	"time"

	"taskmanager/models"
)

var tasks = []models.Task{}
var nextID = 1

func GetAllTasks() []models.Task {
	return tasks
}

func GetTaskById(id string) (models.Task, error) {
	for _, task := range tasks {
		if id == task.ID {
			return task, nil
		}
	}
	return models.Task{}, errors.New("TASK DOESN'T EXIST")
}

func CreateTask(newTask models.Task) models.Task {
	newTask.ID = strconv.Itoa(nextID)
	nextID++ // increment for next task

	// check if duedate is not set
	if newTask.DueDate.IsZero() {
		newTask.DueDate = time.Now()
	}

	tasks = append(tasks, newTask)

	return newTask
}

func UpdateTask(id string, updatedTask models.Task) (models.Task, error) {
	for i, task := range tasks {
		if id == task.ID {
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
			return tasks[i], nil
		}
	}

	return models.Task{}, errors.New("TASK NOT FOUND")

}

func DeleteTask(id string) error {
	for i, task := range tasks {
		if id == task.ID {
			tasks = append(tasks[:1], tasks[i+1:]...)
			return nil
		}
	}
	return errors.New("TASK NOT FOUND")
}

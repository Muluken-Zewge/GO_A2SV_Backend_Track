package usecases

import (
	"context"
	"fmt"
	domain "taskmanager/Domain"
	repositories "taskmanager/Repositories"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type TaskUsecase interface {
	RetrieveAllTasks(ctx context.Context) ([]domain.Task, error)
	RetrieveTaskByID(ctx context.Context, id string) (domain.Task, error)
	CreateTask(ctx context.Context, task domain.Task) (domain.Task, error)
	ModifyTask(ctx context.Context, id string, updatedTask domain.Task) (domain.Task, error)
	RemoveTask(ctx context.Context, id string) error
}

type TaskUsecaseImpl struct {
	taskRepository repositories.TaskRepository
}

// Constructor for dependency injection
func NewTaskUsecase(repo repositories.TaskRepository) TaskUsecase {
	return &TaskUsecaseImpl{
		taskRepository: repo,
	}
}

func (t *TaskUsecaseImpl) RetrieveAllTasks(ctx context.Context) ([]domain.Task, error) {
	tasks, err := t.taskRepository.GetAll(ctx)
	if err != nil {
		return []domain.Task{}, err
	}

	return tasks, nil
}

func (t *TaskUsecaseImpl) RetrieveTaskByID(ctx context.Context, id string) (domain.Task, error) {

	task, err := t.taskRepository.GetByID(ctx, id)
	if err != nil {
		return domain.Task{}, err
	}

	return task, nil
}

func (t *TaskUsecaseImpl) CreateTask(ctx context.Context, task domain.Task) (domain.Task, error) {

	// validat the user input
	if task.Title == "" || task.Description == "" || task.Status == "" {
		return domain.Task{}, fmt.Errorf("%w: title, description and status are required", domain.ErrValidation)
	}

	// assign id
	newId := uuid.New()
	task.ID = newId.String()

	// assign due date if not assigned(it's optional for the user request)
	if task.DueDate.IsZero() {
		task.DueDate = time.Now()
	}

	createdTask, err := t.taskRepository.Create(ctx, task)
	if err != nil {
		return domain.Task{}, err
	}

	return createdTask, nil
}

func (t *TaskUsecaseImpl) ModifyTask(ctx context.Context, id string, updatedTask domain.Task) (domain.Task, error) {

	// build the update document
	updates := bson.M{} // empty unordered map

	// check updated fields
	if updatedTask.Title != "" {
		updates["title"] = updatedTask.Title
	}
	if updatedTask.Description != "" {
		updates["description"] = updatedTask.Description
	}
	if !updatedTask.DueDate.IsZero() {
		updates["due_date"] = updatedTask.DueDate
	}
	if updatedTask.Status != "" {
		updates["status"] = updatedTask.Status
	}

	if len(updates) == 0 {
		t.taskRepository.GetByID(ctx, id)
	}

	updatedTask, err := t.taskRepository.Update(ctx, id, updates)
	if err != nil {
		return domain.Task{}, err
	}

	return updatedTask, nil
}

func (t *TaskUsecaseImpl) RemoveTask(ctx context.Context, id string) error {

	err := t.taskRepository.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

package data

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"taskmanager/models"
)

// TaskService holds the MongoDB connection and collection handle.
type TaskService struct {
	collection *mongo.Collection
}

func NewTaskService(connectionString string, dbName string, collectionName string) (*TaskService, error) {
	// set up a context for connection timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // cancel the context when the function returns

	// set client options
	clientOptions := options.Client().ApplyURI(connectionString)

	// connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// ping the database to verify the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		// Close the client if the ping fails
		client.Disconnect(context.Background())
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	fmt.Println("Successfully connected to MongoDB!")

	// get the collection handle
	collection := client.Database(dbName).Collection(collectionName)

	return &TaskService{
		collection: collection,
	}, nil

}

var nextID = 1

func (t *TaskService) GetAllTasks() ([]models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := t.collection.Find(ctx, primitive.D{})
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks: %w", err)
	}
	defer cursor.Close(ctx) // close the cursor

	var tasks []models.Task
	// decode all the documents to the tasks slice

	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks: %w", err)
	}

	return tasks, nil
}

func (t *TaskService) GetTaskById(id string) (models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var task models.Task

	filter := bson.M{"task_id": id}
	err := t.collection.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.Task{}, errors.New("task not found")
		}
		return models.Task{}, fmt.Errorf("failed to retrieve task: %w", err)
	}

	return task, nil
}

func (t *TaskService) CreateTask(newTask models.Task) (models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	newTask.ID = strconv.Itoa(nextID)
	nextID++ // increment for next task

	// check if duedate is not set
	if newTask.DueDate.IsZero() {
		newTask.DueDate = time.Now()
	}

	_, err := t.collection.InsertOne(ctx, newTask)
	if err != nil {
		return models.Task{}, fmt.Errorf("failed to create task: %w", err)
	}

	return newTask, nil
}

func (t *TaskService) UpdateTask(id string, updatedTask models.Task) (models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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
		t.GetTaskById(id)
	}

	filter := bson.M{"task_id": id}

	// MongoDB query to update all fields inside updates map
	updateQuery := bson.M{"$set": updates}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After) // return the updated document

	var task models.Task

	err := t.collection.FindOneAndUpdate(ctx, filter, updateQuery, opts).Decode(&task)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.Task{}, errors.New("task not found")
		}
		return models.Task{}, fmt.Errorf("failed to update task: %w", err)
	}

	return task, nil
}

func (t *TaskService) DeleteTask(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"task_id": id}
	result, err := t.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	if result.DeletedCount == 0 {
		return errors.New("task not found")
	}

	return nil
}

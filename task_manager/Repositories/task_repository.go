package repositories

import (
	"context"
	"errors"
	"fmt"
	domain "taskmanager/Domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TaskRepository interface {
	GetAll(ctx context.Context) ([]domain.Task, error)
	GetByID(ctx context.Context, id string) (domain.Task, error)
	Create(ctx context.Context, task domain.Task) (domain.Task, error)
	Update(ctx context.Context, id string, updates bson.M) (domain.Task, error)
	Delete(ctx context.Context, id string) error
}

type MongoTaskRepository struct {
	taskCollection *mongo.Collection
}

func NewMongoTaskRepository(client *mongo.Client, dbName string, collectionName string) TaskRepository {
	collection := client.Database(dbName).Collection(collectionName)

	return &MongoTaskRepository{
		taskCollection: collection,
	}
}

func (m *MongoTaskRepository) GetAll(ctx context.Context) ([]domain.Task, error) {

	cursor, err := m.taskCollection.Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks: %w", err)
	}
	defer cursor.Close(ctx) // close the cursor

	var tasks []domain.Task

	// decode all the documents to the tasks slice
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks: %w", err)
	}

	return tasks, nil
}

func (m *MongoTaskRepository) GetByID(ctx context.Context, id string) (domain.Task, error) {

	var task domain.Task

	filter := bson.M{"task_id": id}
	err := m.taskCollection.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Task{}, domain.ErrNotFound
		}
		return domain.Task{}, fmt.Errorf("failed to retrieve task: %w", err)
	}

	return task, nil
}

func (m *MongoTaskRepository) Create(ctx context.Context, task domain.Task) (domain.Task, error) {

	_, err := m.taskCollection.InsertOne(ctx, task)
	if err != nil {
		return domain.Task{}, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}

func (m *MongoTaskRepository) Update(ctx context.Context, id string, updates bson.M) (domain.Task, error) {

	filter := bson.M{"task_id": id}

	// MongoDB query to update all fields inside updates map
	updateQuery := bson.M{"$set": updates}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After) // return the updated document

	var task domain.Task

	err := m.taskCollection.FindOneAndUpdate(ctx, filter, updateQuery, opts).Decode(&task)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Task{}, domain.ErrNotFound
		}
		return domain.Task{}, fmt.Errorf("failed to update task: %w", err)
	}

	return task, nil
}

func (m *MongoTaskRepository) Delete(ctx context.Context, id string) error {

	filter := bson.M{"task_id": id}
	result, err := m.taskCollection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	if result.DeletedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

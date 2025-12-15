package repositoriesintegration

import (
	"context"
	"errors"
	"log"
	"os"
	domain "taskmanager/Domain"
	repositories "taskmanager/Repositories"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TaskRepoTestSuite struct {
	suite.Suite                             // to use suite functionality from testify
	TaskRepo    repositories.TaskRepository // the repo we test
	Client      *mongo.Client               // mongo client
	DBName      string                      // test db name
}

// this function runs once before all tests in the suite
func (suite *TaskRepoTestSuite) SetupSuite() {

	// intialization setup
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("FATAL: MONGO_URI environment variable is not set. Cannot connect to database.")
	}

	suite.DBName = os.Getenv("MONGO_TEST_DB_NAME")
	if suite.DBName == "" {
		suite.DBName = "task_manager_db_test"
	}

	// connect to mongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("FATAL: unable to connect to test database")
	}

	// Ping to ensure connection is live
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("FATAL: MongoDB ping failed: %v", err)
	}

	// assign the dependencies we need as the suite properties
	suite.Client = client

	suite.TaskRepo = repositories.NewMongoTaskRepository(suite.Client, suite.DBName, "tasks")

}

func (suite *TaskRepoTestSuite) TearDownSuite() {

	// CLEANUP: Drop the entire test database to ensure a clean slate.
	suite.Client.Database(suite.DBName).Drop(context.Background())

	// close the connection
	suite.Client.Disconnect(context.Background())
}

// TearDownTest runs after every test in the suite(test isolation)
func (suite *TaskRepoTestSuite) TearDownTest() {

	// We clear all documents from the 'tasks' collection.
	collection := suite.Client.Database(suite.DBName).Collection("tasks")

	// Use an empty filter {} to match all documents.
	_, err := collection.DeleteMany(context.Background(), bson.D{})
	if err != nil {
		log.Printf("Warning: Failed to clear tasks collection after test: %v", err)
	}
}

// a helper to insert a task directly to MongoDB for setup
func (suite *TaskRepoTestSuite) setupTask(taskId, title string) domain.Task {

	// define a task
	task := domain.Task{
		ID:          taskId,
		Title:       title,
		Description: "task description",
		DueDate:     time.Now().Add(24 * time.Hour).Truncate(time.Millisecond),
		Status:      "pending",
	}

	collection := suite.Client.Database(suite.DBName).Collection("tasks")
	_, err := collection.InsertOne(context.Background(), task)
	suite.Require().NoError(err, "failed to insert a task suring setup")

	return task
}

func (suite *TaskRepoTestSuite) TestGetById_Success() {

	// ARRANGE: insert a task to the db
	expectedTask := suite.setupTask("1", "test task")

	// ACT: call the repository method being tested
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	actualTask, err := suite.TaskRepo.GetByID(ctx, expectedTask.ID)

	// ASSERT: check the result and error
	suite.Assert().NoError(err, "GetByID should not return an error for existing task")
	suite.Assert().Equal(expectedTask.ID, actualTask.ID, "Returned task ID should match the expected ID")
	suite.Assert().Equal(expectedTask.Title, actualTask.Title, "Returned task title should match")
}

func (suite *TaskRepoTestSuite) TestGetById_NotFound() {

	// ARRANGE: database is empty, we use an id that definetly won't exist
	nonExistingId := "99"

	// ACT : Try to retrieve the non-existent ID
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := suite.TaskRepo.GetByID(ctx, nonExistingId)

	// ASSERT: check that the error is the expected domain-level error
	// This verifies the repository correctly translates mongo.ErrNoDocuments.
	suite.Assert().Error(err, "GetById should return an error for non-existing task")
	suite.Assert().True(errors.Is(err, domain.ErrNotFound), "Error should be the domain.ErrNotFound")
}

func (suite *TaskRepoTestSuite) TestGetAll_Success() {

	// ARRANGE: Insert three different tasks
	suite.setupTask("1", "Task 1")
	suite.setupTask("2", "Task 2")
	suite.setupTask("3", "Task 3")

	// ACT: retrieve all tasks
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tasks, err := suite.TaskRepo.GetAll(ctx)

	// ASSERT: check the count and error
	suite.Assert().NoError(err, "GetAll should not return an error")
	suite.Assert().Len(tasks, 3, "GetAll should exactly return three tasks")
}

func (suite *TaskRepoTestSuite) TestGetAll_Empty() {

	// ARRANGE: we do nothing(we want empty database)

	// ACT
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tasks, err := suite.TaskRepo.GetAll(ctx)
	// ASSERT
	suite.Assert().NoError(err, "GetAll should return an empty slice, not an error")
	suite.Assert().Empty(tasks, "GetAll should return an empty slice")
}

func (suite *TaskRepoTestSuite) TestCreate_Success() {

	// ARRANGE: define the task to be inserted
	newTask := domain.Task{
		ID:          "1",
		Title:       "test task",
		Description: "test task description",
		DueDate:     time.Now().Add(time.Hour * 24).Truncate(time.Millisecond),
		Status:      "pending",
	}

	// ACT: call the repo method to create a task
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	createdTask, err := suite.TaskRepo.Create(ctx, newTask)

	// ASSERT: 1. check the repository call result
	suite.Require().NoError(err, "Create should not return error on success")
	suite.Assert().Equal(createdTask.Title, newTask.Title, "the returned task should match the input task")

	// ASSERT: 2. verify the document actually exists in the database
	collection := suite.Client.Database(suite.DBName).Collection("tasks")

	var retrievedTask domain.Task

	err = collection.FindOne(context.Background(), bson.M{"task_id": newTask.ID}).Decode(&retrievedTask)

	suite.Require().NoError(err, "Direct database query should find the created task")
	suite.Assert().Equal(retrievedTask.ID, newTask.ID, "The ID in the database should match the created task ID")
}

func (suite *TaskRepoTestSuite) TestUpdate_Success() {

	// ARRANGE: insert a task to be updated and create the update map
	intialTask := suite.setupTask("1", "test task")

	updates := bson.M{
		"title":       "updated test task",
		"description": "updated test task description",
	}

	// ACT: call the repo method
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	updatedTask, err := suite.TaskRepo.Update(ctx, intialTask.ID, updates)

	// ASSERT: 1. check the repository call result
	suite.Require().NoError(err, "Update shouldn't return an error on success")
	suite.Require().Equal(updatedTask.Title, updates["title"], "title should be updated")
	suite.Require().Equal(updatedTask.Description, updates["description"], "title should be updated")

	// ASSERT: 2. verify the document is updated in the database
	collection := suite.Client.Database(suite.DBName).Collection("tasks")
	var dbCheckTask domain.Task
	err = collection.FindOne(context.Background(), bson.M{"task_id": intialTask.ID}).Decode(&dbCheckTask)

	suite.Require().NoError(err, "error shouldn't be returned for a database check")
	suite.Assert().Equal(dbCheckTask.Title, updates["title"], "Database check confirms title update")
}

func (suite *TaskRepoTestSuite) TestUpdate_NotFound() {

	// ARRANGE: no set up needed

	// define the update map
	updates := bson.M{
		"title": "non-existing task title",
	}

	// ACT
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := suite.TaskRepo.Update(ctx, "1", updates)

	// ASSERT
	suite.Require().Error(err, "Update should return error on non-existing task")
	suite.True(errors.Is(err, domain.ErrNotFound), "Error should be the domain.ErrNotFound")
}

func (suite *TaskRepoTestSuite) TestDelete_Success() {

	// ARRANGE: insert task to be deleted
	intialTask := suite.setupTask("1", "test task")

	// ACT
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := suite.TaskRepo.Delete(ctx, intialTask.ID)

	// ASSERT: 1. check the repo call result
	suite.Require().NoError(err, "Delete shouldn't return an error for success")

	// ASSERT: 2. confirm the task no longer exists in the database
	collection := suite.Client.Database(suite.DBName).Collection("tasks")
	var dbCheckTask domain.Task

	err = collection.FindOne(context.Background(), bson.M{"task_id": intialTask.ID}).Decode(&dbCheckTask)

	suite.Assert().True(errors.Is(err, domain.ErrNotFound), "Direct query should confirm the task was deleted")
}

func (suite *TaskRepoTestSuite) TestDelete_NotFound() {

	// ARRANGE: no set up needed

	// ACT
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := suite.TaskRepo.Delete(ctx, "1")

	// ASSERT
	suite.Require().Error(err, "Delete should return error on non-existing task")
	suite.True(errors.Is(err, domain.ErrNotFound), "Error should be the domain.ErrNotFound")
}

// This function is the entry point for the 'go test' command.
func TestTaskRepoSuite(t *testing.T) {
	// looks for the Test* methods in TaskRepoTestSuite
	suite.Run(t, new(TaskRepoTestSuite))
}

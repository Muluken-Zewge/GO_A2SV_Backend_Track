package usecases_test

import (
	"context"
	"errors"
	"testing"

	domain "taskmanager/Domain"
	"taskmanager/Tests/mocks"
	usecases "taskmanager/Usecases"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
)

type TaskUsecaseTestSuite struct {
	suite.Suite
	mockRepo *mocks.MockTaskRepository
	usecase  usecases.TaskUsecase
}

func (suite *TaskUsecaseTestSuite) SetupTest() {
	// Initialize the mock and the usecase before each test
	suite.mockRepo = new(mocks.MockTaskRepository)
	suite.usecase = usecases.NewTaskUsecase(suite.mockRepo)
}

// --- 1. Test CreateTask ---

func (suite *TaskUsecaseTestSuite) TestCreateTask_Success() {
	ctx := context.TODO()
	inputTask := domain.Task{
		Title:       "Test Task",
		Description: "Testing logic",
		Status:      "pending",
	}

	// EXPECT: The repository to be called with a task that has a generated ID
	suite.mockRepo.EXPECT().
		Create(ctx, mock.MatchedBy(func(t domain.Task) bool {
			return t.Title == inputTask.Title && t.ID != "" // Check if ID was assigned
		})).
		Return(domain.Task{ID: "generated-uuid", Title: "Test Task"}, nil)

	result, err := suite.usecase.CreateTask(ctx, inputTask)

	suite.NoError(err)
	suite.Equal("generated-uuid", result.ID)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *TaskUsecaseTestSuite) TestCreateTask_ValidationError() {
	ctx := context.TODO()
	invalidTask := domain.Task{Title: ""} // Missing description and status

	result, err := suite.usecase.CreateTask(ctx, invalidTask)

	suite.Error(err)
	suite.True(errors.Is(err, domain.ErrValidation))
	suite.Empty(result)
	// Ensure the repository was NEVER called
	suite.mockRepo.AssertNotCalled(suite.T(), "Create", mock.Anything, mock.Anything)
}

// --- 2. Test RetrieveTaskByID ---

func (suite *TaskUsecaseTestSuite) TestRetrieveTaskByID_Success() {
	ctx := context.TODO()
	taskID := "123"
	expectedTask := domain.Task{ID: taskID, Title: "Existing Task"}

	suite.mockRepo.EXPECT().GetByID(ctx, taskID).Return(expectedTask, nil)

	result, err := suite.usecase.RetrieveTaskByID(ctx, taskID)

	suite.NoError(err)
	suite.Equal(expectedTask, result)
}

func (suite *TaskUsecaseTestSuite) TestRetrieveTaskByID_NotFound() {
	ctx := context.TODO()
	suite.mockRepo.EXPECT().GetByID(ctx, "999").Return(domain.Task{}, domain.ErrNotFound)

	_, err := suite.usecase.RetrieveTaskByID(ctx, "999")

	suite.Error(err)
	suite.True(errors.Is(err, domain.ErrNotFound))
}

// --- 3. Test ModifyTask ---

func (suite *TaskUsecaseTestSuite) TestModifyTask_Success() {
	ctx := context.TODO()
	id := "1"
	updatedFields := domain.Task{
		Title: "New Title",
	}

	// The logic inside ModifyTask builds a bson.M{"title": "New Title"}
	expectedUpdates := bson.M{"title": "New Title"}

	suite.mockRepo.EXPECT().
		Update(ctx, id, expectedUpdates).
		Return(domain.Task{ID: id, Title: "New Title"}, nil)

	result, err := suite.usecase.ModifyTask(ctx, id, updatedFields)

	suite.NoError(err)
	suite.Equal("New Title", result.Title)
}

// --- 4. Test RemoveTask ---

func (suite *TaskUsecaseTestSuite) TestRemoveTask_Success() {
	ctx := context.TODO()
	id := "delete-me"

	suite.mockRepo.EXPECT().Delete(ctx, id).Return(nil)

	err := suite.usecase.RemoveTask(ctx, id)

	suite.NoError(err)
}

func TestTaskUsecaseTestSuite(t *testing.T) {
	suite.Run(t, new(TaskUsecaseTestSuite))
}

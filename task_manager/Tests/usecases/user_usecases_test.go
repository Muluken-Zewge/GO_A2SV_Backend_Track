package usecases_test

import (
	"context"
	"errors"
	"os"
	"testing"

	domain "taskmanager/Domain"
	infrastructure "taskmanager/Infrastructure"
	"taskmanager/Tests/mocks"
	usecases "taskmanager/Usecases"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type UserUsecaseTestSuite struct {
	suite.Suite
	mockRepo *mocks.MockUserRepository
	usecase  usecases.UserUsecase
}

func (suite *UserUsecaseTestSuite) SetupTest() {
	suite.mockRepo = new(mocks.MockUserRepository)
	suite.usecase = usecases.NewUserUsecase(suite.mockRepo)

	// Set JWT_SECRET for infrastructure.GenerateJWT
	os.Setenv("JWT_SECRET", "test_secret")
}

func (suite *UserUsecaseTestSuite) TearDownTest() {
	os.Unsetenv("JWT_SECRET")
}

// --- 1. Test RegisterUser ---

func (suite *UserUsecaseTestSuite) TestRegisterUser_Success_FirstUserIsAdmin() {
	ctx := context.TODO()
	userName := "admin_user"
	password := "securepassword"

	// 1. Mock Username Availability
	suite.mockRepo.EXPECT().IsUsernameAvailable(ctx, userName).Return(nil)

	// 2. Mock Database Empty Check (returns true, so user should be Admin)
	suite.mockRepo.EXPECT().IsDatabaseEmpty(ctx).Return(true, nil)

	// 3. Mock SaveUser
	// We use MatchBy to ensure the role was correctly set to Admin logic-wise
	suite.mockRepo.EXPECT().
		SaveUser(ctx, mock.MatchedBy(func(u domain.User) bool {
			return u.UserName == userName && u.Role == domain.RoleAdmin
		})).
		Return(domain.User{UserName: userName, Role: domain.RoleAdmin}, nil)

	result, err := suite.usecase.RegisterUser(ctx, userName, password)

	suite.NoError(err)
	suite.Equal(domain.RoleAdmin, result.Role)
}

func (suite *UserUsecaseTestSuite) TestRegisterUser_Fail_ShortPassword() {
	ctx := context.TODO()
	_, err := suite.usecase.RegisterUser(ctx, "user", "123") // 3 chars

	suite.Error(err)
	suite.Contains(err.Error(), "password should be at least 4 characters")
	suite.mockRepo.AssertNotCalled(suite.T(), "IsUsernameAvailable", mock.Anything, mock.Anything)
}

func (suite *UserUsecaseTestSuite) TestRegisterUser_Fail_UsernameExists() {
	ctx := context.TODO()
	userName := "taken_name"

	suite.mockRepo.EXPECT().IsUsernameAvailable(ctx, userName).Return(domain.ErrAleadyExists)

	_, err := suite.usecase.RegisterUser(ctx, userName, "password")

	suite.Error(err)
	suite.True(errors.Is(err, domain.ErrAleadyExists))
}

// --- 2. Test AuthenticateUser ---

func (suite *UserUsecaseTestSuite) TestAuthenticateUser_Success() {
	ctx := context.TODO()
	userName := "john_doe"
	password := "secret123"

	// Create a real hash so infrastructure.ComparePassword succeeds
	hashedPassword, _ := infrastructure.HashPassword(password)
	userID := uuid.New().String()

	// 1. Mock user existence check
	suite.mockRepo.EXPECT().
		DoesUserExist(ctx, userName).
		Return(userID, hashedPassword, domain.RoleUser, nil)

	// ACT
	token, err := suite.usecase.AuthenticateUser(ctx, userName, password)

	// ASSERT
	suite.NoError(err)
	suite.NotEmpty(token, "Should return a valid JWT string")
}

func (suite *UserUsecaseTestSuite) TestAuthenticateUser_Fail_WrongPassword() {
	ctx := context.TODO()
	userName := "john_doe"

	correctPassword := "correct_one"
	wrongPassword := "wrong_one"
	hashedPassword, _ := infrastructure.HashPassword(correctPassword)

	suite.mockRepo.EXPECT().
		DoesUserExist(ctx, userName).
		Return(uuid.New().String(), hashedPassword, domain.RoleUser, nil)

	// ACT
	token, err := suite.usecase.AuthenticateUser(ctx, userName, wrongPassword)

	// ASSERT
	suite.Error(err)
	suite.Empty(token)
	suite.True(errors.Is(err, domain.ErrValidation), "Should return validation error on password mismatch")
}

// --- 3. Test PromoteUser ---

func (suite *UserUsecaseTestSuite) TestPromoteUser_Success() {
	ctx := context.TODO()
	targetID := uuid.New().String()

	suite.mockRepo.EXPECT().
		PromoteUser(ctx, targetID).
		Return(domain.User{ID: uuid.MustParse(targetID), Role: domain.RoleAdmin}, nil)

	result, err := suite.usecase.PromoteUser(ctx, targetID)

	suite.NoError(err)
	suite.Equal(domain.RoleAdmin, result.Role)
}

func TestUserUsecaseTestSuite(t *testing.T) {
	suite.Run(t, new(UserUsecaseTestSuite))
}

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

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepoTestSuite struct {
	suite.Suite                             // to use suite functionality from testify
	UserRepo    repositories.UserRepository // the repo we test
	Client      *mongo.Client               // mongo client
	DBName      string                      // test db name
}

func (suite *UserRepoTestSuite) SetupSuite() {

	// Load variables from .env file
	if err := godotenv.Load("../../config/.env"); err != nil {
		log.Println("Note: No .env file found, relying on system environment variables.")
	}

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

	suite.UserRepo = repositories.NewMongoUserRepository(suite.Client, suite.DBName, "users")
}

func (suite *UserRepoTestSuite) TearDownSuite() {

	// CLEANUP: Drop the entire test database to ensure a clean slate.
	suite.Client.Database(suite.DBName).Drop(context.Background())

	// close the connection
	suite.Client.Disconnect(context.Background())
}

func (suite *UserRepoTestSuite) TearDownTest() {

	// We clear all documents from the 'tasks' collection.
	collection := suite.Client.Database(suite.DBName).Collection("users")

	// Use an empty filter {} to match all documents.
	_, err := collection.DeleteMany(context.Background(), bson.D{})
	if err != nil {
		log.Printf("Warning: Failed to clear tasks collection after test: %v", err)
	}
}

// helper to insert a user to the test database
func (suite *UserRepoTestSuite) setupUser(userName, password string, role int) domain.User {

	user := domain.User{
		ID:             uuid.New(),
		UserName:       userName,
		HashedPassword: password, // not actually hashed, just test data
		Role:           domain.UserRole(role),
	}

	collection := suite.Client.Database(suite.DBName).Collection("users")
	_, err := collection.InsertOne(context.Background(), user)
	suite.Require().NoError(err, "Failed to insert user during setup")
	return user
}

func TestUserRepoSuite(t *testing.T) {
	// looks for the Test* methods in TaskRepoTestSuite
	suite.Run(t, new(UserRepoTestSuite))
}

func (suite *UserRepoTestSuite) TestIsUserNameAvailable_Available() {
	// ARRANGE: we do notthing here, the database is empty

	// ACT
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := suite.UserRepo.IsUsernameAvailable(ctx, "uniqeusername")

	// ASSERT
	suite.Require().NoError(err, "user name should be available")
}

func (suite *UserRepoTestSuite) TestIsUserNameAvailable_NotAvailable() {

	//ARRANGE: insert a user
	suite.setupUser("uniqueuser", "password", 0)

	// ACT
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := suite.UserRepo.IsUsernameAvailable(ctx, "uniqueuser")

	// ASSERT
	suite.Require().Error(err, "username check should return error for existing user name")
	suite.Assert().True(errors.Is(err, domain.ErrAleadyExists), "error should be domain.ErrAleadyExists")
}

func (suite *UserRepoTestSuite) TestIsDatabaseEmpty_Empty() {
	// ARRANGE: we do nothing, the database is empty

	// ACT
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	isEmpty, err := suite.UserRepo.IsDatabaseEmpty(ctx)

	// ASSERT
	suite.Require().NoError(err)
	suite.Assert().True(isEmpty, "IsDatabaseEmpty should return true for empty database")
}

func (suite *UserRepoTestSuite) TestIsDatabaseEmpty_NotEmpty() {

	// ARRANGE
	suite.setupUser("username", "password", 0)

	// ACT
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	isEmpty, err := suite.UserRepo.IsDatabaseEmpty(ctx)

	// ASSERT
	suite.Require().NoError(err)
	suite.Assert().False(isEmpty, "IsDatabaseEmpty should return false for non-empty database")
}

func (suite *UserRepoTestSuite) TestSaveUser_success() {

	// ARRANGE: create a user to save
	userToSave := domain.User{
		ID:             uuid.New(),
		UserName:       "username",
		HashedPassword: "password",
		Role:           domain.RoleUser,
	}

	// ACT
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	savedUser, err := suite.UserRepo.SaveUser(ctx, userToSave)

	// ASSERT: 1. check the repo resonse
	suite.Require().NoError(err, "SaveUser shouldn't return error")
	suite.Assert().Equal(savedUser.UserName, userToSave.UserName, "savedUser should have the same username")

	// ASSERT: 2. check if the user is actually saved in the database
	collection := suite.Client.Database(suite.DBName).Collection("users")
	count, err := collection.CountDocuments(context.Background(), bson.M{"user_name": userToSave.UserName})

	suite.Require().NoError(err)
	suite.Assert().Equal(int(count), 1, "One user should be found in the database")
}

func (suite *UserRepoTestSuite) TestDoesUserExist_Exist() {

	// ARRANGE
	insertedUser := suite.setupUser("username", "password", 1)

	// ACT
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	id, _, role, err := suite.UserRepo.DoesUserExist(ctx, insertedUser.UserName)

	// ASSERT
	suite.Require().NoError(err)
	suite.Assert().Equal(id, insertedUser.ID.String(), "user id should match")
	suite.Assert().Equal(role, insertedUser.Role, "user role should match")
}

func (suite *UserRepoTestSuite) TestDoesUserExist_NotFound() {

	// ARRANGE: we do nothing

	// ACT
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, _, _, err := suite.UserRepo.DoesUserExist(ctx, "username")

	// ASSERT
	suite.Require().Error(err, "DoesUserExist should return an error for non-existent user")
	suite.Assert().True(errors.Is(err, domain.ErrNotFound), "error should be domain.ErrNotFound")
}

func (suite *UserRepoTestSuite) TestPromoteUser_Success() {
	// ARRANGE: add a user with RoleUser
	insertedUser := suite.setupUser("username", "password", 0)

	// ACT
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	promotedUser, err := suite.UserRepo.PromoteUser(ctx, insertedUser.ID.String())

	// ASSERT: check the repo response
	suite.Require().NoError(err)
	suite.Assert().Equal(promotedUser.Role, domain.RoleAdmin, "user role should be updated to RoleAdmin")

	// ASSERT 2: Verify the document is updated in the database.
	collection := suite.Client.Database(suite.DBName).Collection("users")
	var dbCheckUser domain.User
	err = collection.FindOne(context.Background(), bson.M{"user_id": insertedUser.ID}).Decode(&dbCheckUser)

	suite.Require().NoError(err)
	suite.Assert().Equal(domain.RoleAdmin, dbCheckUser.Role, "Database check confirms role update")
}

func (suite *UserRepoTestSuite) TestPromoteUser_NotFound() {
	// ARRANGE: Use a valid, but non-existent UUID.
	nonExistentUUID := uuid.New().String()

	// ACT: Try to promote a non-existent user.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := suite.UserRepo.PromoteUser(ctx, nonExistentUUID)

	// ASSERT: Should return domain.ErrNotFound.
	suite.Require().Error(err)
	suite.Assert().True(errors.Is(err, domain.ErrNotFound), "Error should be domain.ErrNotFound")
}

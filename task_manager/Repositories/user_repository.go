package repositories

import (
	"context"
	"errors"
	"fmt"
	domain "taskmanager/Domain"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository interface {
	IsUsernameAvailable(ctx context.Context, userName string) error
	IsDatabaseEmpty(ctx context.Context) (bool, error)
	SaveUser(ctx context.Context, user domain.User) (domain.User, error)
	DoesUserExist(ctx context.Context, userName string) (string, string, domain.UserRole, error)
	PromoteUser(ctx context.Context, userId string) (domain.User, error)
}

type MongoUserRepository struct {
	userCollection *mongo.Collection
}

func NewMongoUserRepository(client *mongo.Client, dbName string, collectionName string) UserRepository {
	collection := client.Database(dbName).Collection(collectionName)

	return &MongoUserRepository{
		userCollection: collection,
	}
}

func (m *MongoUserRepository) IsUsernameAvailable(ctx context.Context, userName string) error {

	filter := bson.M{"user_name": userName}

	// a variable to store the result
	var existingUser struct{}

	// use a projection to retrieve only the id
	options := options.FindOne().SetProjection(bson.D{{Key: "_id", Value: 1}})

	err := m.userCollection.FindOne(ctx, filter, options).Decode(&existingUser)
	if err == nil {
		// If NO error, the document was found.
		return domain.ErrAleadyExists
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("error checking username uniqueness: %w", err)
	}

	return nil
}

func (m *MongoUserRepository) IsDatabaseEmpty(ctx context.Context) (bool, error) {

	// check if the database is empty or not
	count, err := m.userCollection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return false, fmt.Errorf("error checking collection count: %w", err)
	}
	if count == 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func (m *MongoUserRepository) SaveUser(ctx context.Context, user domain.User) (domain.User, error) {
	// save user to the database
	_, err := m.userCollection.InsertOne(ctx, user)
	if err != nil {
		return domain.User{}, fmt.Errorf("error registering user: %w", err)
	}

	return user, nil
}

func (m *MongoUserRepository) DoesUserExist(ctx context.Context, userName string) (string, string, domain.UserRole, error) {

	// check if the user name exists
	filter := bson.M{"user_name": userName}

	var existingUser struct {
		SavedPassword string          `bson:"hashed_password"`
		UserId        uuid.UUID       `bson:"user_id"`
		Role          domain.UserRole `bson:"role"`
	}

	options := options.FindOne().SetProjection(bson.D{{Key: "user_id", Value: 1}, {Key: "hashed_password", Value: 1}, {Key: "role", Value: 1}})

	err := m.userCollection.FindOne(ctx, filter, options).Decode(&existingUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", "", domain.RoleUser, domain.ErrNotFound
		}
		return "", "", domain.RoleUser, fmt.Errorf("error checking user name: %w", err)
	}

	return existingUser.UserId.String(), existingUser.SavedPassword, existingUser.Role, nil
}

func (m *MongoUserRepository) PromoteUser(ctx context.Context, userId string) (domain.User, error) {

	// prepare query components
	parsedUUID, err := uuid.Parse(userId)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed to parse string id to uuid type")
	}
	filter := bson.M{"user_id": parsedUUID}

	updateData := bson.M{}
	updateData["role"] = domain.RoleAdmin

	updateQuery := bson.M{"$set": updateData}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	// execute database command
	var user domain.User
	err = m.userCollection.FindOneAndUpdate(ctx, filter, updateQuery, opts).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, fmt.Errorf("failed to update user status: %w", err)
	}

	return user, nil
}

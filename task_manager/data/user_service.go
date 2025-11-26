package data

import (
	"context"
	"errors"
	"fmt"
	"os"
	"taskmanager/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userCollection *mongo.Collection
}

func NewUserService(client *mongo.Client, dbName string, collectionName string) *UserService {
	// get collection handle
	collection := client.Database(dbName).Collection(collectionName)

	return &UserService{
		userCollection: collection,
	}
}

func (us *UserService) RegisterUser(userCredential models.Credentials) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// define the database user struct
	var newUser models.User

	// validate the password
	if len(userCredential.Password) < 4 {
		return models.User{}, errors.New("password should be at least 4 characters")
	}
	// check if the user name is unique(doesn't exist in the collection)
	filter := bson.M{"user_name": userCredential.UserName}

	// a variable to store the result
	var existingUser struct{}

	// use a projection to retrieve only the id
	options := options.FindOne().SetProjection(bson.D{{Key: "_id", Value: 1}})

	err := us.userCollection.FindOne(ctx, filter, options).Decode(&existingUser)
	if err == nil {
		// If NO error, the document was found.
		return models.User{}, errors.New("user name already exist")
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return models.User{}, fmt.Errorf("error checking username uniqueness: %w", err)
	}

	// assign username to database user
	newUser.UserName = userCredential.UserName

	// hash the password and assign to database user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userCredential.Password), bcrypt.DefaultCost)

	if err != nil {
		return models.User{}, fmt.Errorf("failed to hash password: %w", err)
	}
	newUser.HashedPassword = string(hashedPassword)

	// assign user id(uuid)
	var newID = uuid.New()
	newUser.ID = newID

	//if there's no user, assign the first user admin role and the next ones just user
	count, err := us.userCollection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return models.User{}, fmt.Errorf("error checking collection count: %w", err)
	}
	if count == 0 {
		newUser.Role = models.RoleAdmin
	} else {
		newUser.Role = models.RoleUser
	}

	// make the db call
	_, err = us.userCollection.InsertOne(ctx, newUser)
	if err != nil {
		return models.User{}, fmt.Errorf("error registering user: %w", err)
	}

	return newUser, nil
}

func (us *UserService) AuthenticateUser(userCredential models.Credentials) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// check if the user name exists
	filter := bson.M{"user_name": userCredential.UserName}

	var existingUser struct {
		SavedPassword string          `bson:"hashed_password"`
		UserId        uuid.UUID       `bson:"user_id"`
		Role          models.UserRole `bson:"role"`
	}

	options := options.FindOne().SetProjection(bson.D{{Key: "user_id", Value: 1}, {Key: "hashed_password", Value: 1}, {Key: "role", Value: 1}})

	err := us.userCollection.FindOne(ctx, filter, options).Decode(&existingUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", errors.New("invalid credential")
		}
		return "", fmt.Errorf("error checking user name: %w", err)
	}

	// check if password is correct
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.SavedPassword), []byte(userCredential.Password))
	if err != nil {
		return "", errors.New("invalid credential")
	}

	// generate jwt token
	jwtSecret := os.Getenv("JWT_SECRET") // get jwt secret from env varaiable

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   existingUser.UserId,
		"user_name": userCredential.UserName,
		"role":      existingUser.Role,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	})

	// sign the token with the secret key
	jwtToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}

	return jwtToken, nil
}

func (us *UserService) PromoteUser(userId string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// prepare query components
	parsedUUID, err := uuid.Parse(userId)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to parse string id to uuid type")
	}
	filter := bson.M{"user_id": parsedUUID}

	updateData := bson.M{}
	updateData["role"] = models.RoleAdmin

	updateQuery := bson.M{"$set": updateData}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	// execute database command
	var user models.User
	err = us.userCollection.FindOneAndUpdate(ctx, filter, updateQuery, opts).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, fmt.Errorf("failed to update user status: %w", err)
	}

	return user, nil
}

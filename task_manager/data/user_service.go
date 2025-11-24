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

func (us *UserService) RegisterUser(user models.User) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// validate the password
	if len(user.Password) < 4 {
		return models.User{}, errors.New("password should be at least 4 characters")
	}
	// check if the user name is unique(doesn't exist in the collection)
	filter := bson.M{"user_name": user.UserName}

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

	// hash the password and assign
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	if err != nil {
		return models.User{}, fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	// assign user id(uuid)
	var newID = uuid.New()
	user.ID = newID

	//if there's no user, assign the first user admin role and the next ones just user
	count, err := us.userCollection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return models.User{}, fmt.Errorf("error checking collection count: %w", err)
	}
	if count == 0 {
		user.Role = models.RoleAdmin
	} else {
		user.Role = models.RoleAdmin
	}

	// make the db call
	_, err = us.userCollection.InsertOne(ctx, user)
	if err != nil {
		return models.User{}, fmt.Errorf("error registering user: %w", err)
	}

	return user, nil
}

func (us *UserService) AuthenticateUser(user models.User) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// check if the user name exists
	filter := bson.M{"user_name": user.UserName}

	var existingUser struct {
		SavedPassword string          `bson:"password"`
		UserId        uuid.UUID       `bson:"user_id"`
		Role          models.UserRole `bson:"role"`
	}

	options := options.FindOne().SetProjection(bson.D{{Key: "user_id", Value: 1}, {Key: "password", Value: 1}})

	err := us.userCollection.FindOne(ctx, filter, options).Decode(&existingUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", errors.New("invalid credential")
		}
		return "", fmt.Errorf("error checking user name: %w", err)
	}

	// check if password is correct
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.SavedPassword), []byte(user.Password))
	if err != nil {
		return "", errors.New("invalid credential")
	}

	// generate jwt token
	jwtSecret := os.Getenv("JWT_SECRET") // get jwt secret from env varaiable

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   existingUser.UserId,
		"user_name": user.UserName,
		"role":      existingUser.Role,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	})

	// sign the token with the secret key
	jwtToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}

	return jwtToken, nil
}

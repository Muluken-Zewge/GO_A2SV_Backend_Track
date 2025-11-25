package main

import (
	"context"
	"log"
	"os"
	"taskmanager/data"
	"taskmanager/router"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Note: No .env file found, relying on system environment variables.")
	}

	// Retrieve all necessary configuration from environment variables
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB_NAME")
	taskCollectionName := os.Getenv("MONGO_TASK_COLLECTION")
	userCollectionName := os.Getenv("MONGO_USER_COLLECTION")

	// Critical Validation: Ensure the URI is set
	if mongoURI == "" {
		log.Fatal("FATAL: MONGO_URI environment variable is not set. Cannot connect to database.")
	}

	// 3. Fallback/Validation for DB/Collection
	if dbName == "" {
		dbName = "task_db"
		log.Println("Using default database name: task_db")
	}
	if taskCollectionName == "" {
		taskCollectionName = "tasks"
		log.Println("Using default task collection name: tasks")
	}

	if userCollectionName == "" {
		userCollectionName = "users"
		log.Println("Using default user collection name: users")
	}

	/// ---CREAT MONGODB CONNECTION ---

	// set up a context for connection timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // cancel the context when the function returns

	// set client options
	clientOptions := options.Client().ApplyURI(mongoURI)

	// connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("FATAL: unable to connect to database")
	}

	// Ensure the client is closed when main() exits or panics
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("FATAL: Error disconnecting MongoDB client: %v", err)
		}
	}()

	// Ping the primary database to verify connection and credentials
	err = client.Ping(ctx, nil)
	if err != nil {
		// Close the client gracefully if the ping fails
		client.Disconnect(context.Background())
		log.Fatalf("FATAL: Failed to ping MongoDB: %v", err)
	}

	log.Println("Successfully connected to MongoDB Atlas.")

	// intialize task service
	taskService := data.NewTaskService(client, dbName, taskCollectionName)

	// intialize user service
	userService := data.NewUserService(client, dbName, userCollectionName)

	r := router.SetupRouter(taskService, userService)

	log.Println("Server starting on port 8080...")

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}

package main

import (
	"log"
	"os"
	"taskmanager/data"
	"taskmanager/router"

	"github.com/joho/godotenv"
)

func main() {
	// Load variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Note: No .env file found, relying on system environment variables.")
	}

	// Retrieve all necessary configuration from environment variables
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB_NAME")
	collectionName := os.Getenv("MONGO_COLLECTION_NAME")

	// Critical Validation: Ensure the URI is set
	if mongoURI == "" {
		log.Fatal("FATAL: MONGO_URI environment variable is not set. Cannot connect to database.")
	}

	// 3. Fallback/Validation for DB/Collection
	if dbName == "" {
		dbName = "task_db"
		log.Println("Using default database name: task_db")
	}
	if collectionName == "" {
		collectionName = "tasks"
		log.Println("Using default collection name: tasks")
	}

	// intialize task service
	taskService, err := data.NewTaskService(mongoURI, dbName, collectionName)
	if err != nil {
		log.Fatalf("Service initialization failed: %v", err)
	}
	r := router.SetupRouter(taskService)

	log.Println("Server starting on port 8080...")

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}

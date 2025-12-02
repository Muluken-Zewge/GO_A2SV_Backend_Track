package domain

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Task struct {
	// MongoDB's navtive _id(internal, private field)
	id primitive.ObjectID `bson:"_id,omitempty" json:"-"` // igonre it in the json
	// custom id for task identification
	ID          string    `json:"id" bson:"task_id"`
	Title       string    `json:"title" bson:"title"`
	Description string    `json:"description" bson:"description"`
	DueDate     time.Time `json:"due_date" bson:"due_date"`
	Status      string    `json:"status" bson:"status"`
}

// define user role enum
type UserRole int

const (
	RoleUser UserRole = iota
	RoleAdmin
)

// The main struct stored in the database
type User struct {
	mongoID        primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	ID             uuid.UUID          `bson:"user_id" json:"id"`
	UserName       string             `bson:"user_name" json:"user_name"`
	HashedPassword string             `bson:"hashed_password" json:"-"`
	Role           UserRole           `bson:"role" json:"role"`
}

// Used only for binding credentials from the client's request body
type Credentials struct {
	UserName string `json:"user_name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

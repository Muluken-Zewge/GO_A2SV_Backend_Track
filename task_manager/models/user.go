package models

import (
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

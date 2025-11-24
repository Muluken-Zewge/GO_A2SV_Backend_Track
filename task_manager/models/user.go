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

// define the user struct
type User struct {
	mongoID  primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	ID       uuid.UUID          `bson:"user_id" json:"id"`
	UserName string             `bson:"user_name" json:"user_name"`
	Password string             `bson:"password" json:"-"`
	Role     UserRole           `bson:"role" josn:"role"`
}

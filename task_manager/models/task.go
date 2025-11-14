package models

import (
	"time"

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

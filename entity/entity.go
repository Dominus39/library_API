package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Username string             `json:"username"`
	Password string             `json:"password"`
}

type Book struct {
	ID            primitive.ObjectID `json:"_id" bson:"_id, omitempty"`
	Title         string             `json:"title"`
	Author        string             `json:"author"`
	PublishedDate time.Time          `json:"published_date" bson:"published_date"`
	Status        string             `json:"status"`
}

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

type BorrowedBooks struct {
	ID           primitive.ObjectID `json:"_id" bson:"_id, omitempty"`
	BookID       string             `json:"book_id" bson:"book_id"`
	UserID       string             `json:"user_id" bson:"user_id"`
	BorrowedDate string             `json:"borrowed_date" bson:"borrowed_date"`
	ReturnDate   string             `json:"return_date" bsong:"return_date"`
}

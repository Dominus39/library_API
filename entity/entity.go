package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       primitive.ObjectID `json:"user_id" bson:"user_id"`
	Username string             `json:"username"`
	Password string             `json:"password"`
}

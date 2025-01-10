package main

import (
	"context"
	"graded-challange-2-Dominus39/pb"

	"go.mongodb.org/mongo-driver/mongo"
)

type BookRentalServiceServer struct {
	pb.UnimplementedBookRentalServiceServer
	usersCollection *mongo.Collection
}

func (s *BookRentalServiceServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {

}

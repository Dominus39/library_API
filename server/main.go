package main

import (
	"context"
	"fmt"
	"gc2-yugo/config"
	"gc2-yugo/entity"
	"gc2-yugo/pb"
	"log"
	"net"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type BookRentalServiceServer struct {
	pb.UnimplementedBookRentalServiceServer
	usersCollection *mongo.Collection
}

func (s *BookRentalServiceServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	var existingUser entity.User
	err := s.usersCollection.FindOne(ctx, bson.M{"username": req.Username}).Decode(&existingUser)
	if err != nil {
		return nil, status.Errorf(codes.AlreadyExists, "username already exists")
	}

	newUser := entity.User{
		ID:       primitive.NewObjectID(),
		Username: req.Username,
		Password: req.Password,
	}

	_, err = s.usersCollection.InsertOne(ctx, newUser)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	fmt.Printf("User registered: %s", req.Username)

	return &pb.RegisterUserResponse{
		Message: "User registered successfully",
		UserId:  newUser.ID.Hex(),
	}, nil
}

func UnaryAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx, err := AuthInterceptor(ctx)
	if err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func AuthInterceptor(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		fmt.Println("No metadata found")
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized")
	}

	token := md["authorization"]
	if len(token) == 0 {
		fmt.Println("Invalid or missing token")
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized")
	}

	fmt.Println("Token validated successfully")
	return ctx, nil
}

func main() {
	ctx := context.Background()

	usersCollection, err := config.ConnectionDatabaseUsers(ctx)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(UnaryAuthInterceptor),
	)
	bookRentalService := &BookRentalServiceServer{
		usersCollection: usersCollection,
	}

	pb.RegisterBookRentalServiceServer(grpcServer, bookRentalService)

	listen, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("Server is running on port 50051 . . .")

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

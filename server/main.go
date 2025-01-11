package main

import (
	"context"
	"fmt"
	"gc2-yugo/config"
	"gc2-yugo/entity"
	"gc2-yugo/pb"
	"log"
	"net"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
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
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "username already exists")
	}
	if err != mongo.ErrNoDocuments {
		return nil, status.Errorf(codes.Internal, "failed to check existing user: %v", err)
	}

	newUser := entity.User{
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
	}, nil
}

func (s *BookRentalServiceServer) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	var user entity.User

	err := s.usersCollection.FindOne(ctx, bson.M{"username": req.Username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, "username not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to fetch user: %v", err)
	}

	if user.Password != req.Password {
		return nil, status.Errorf(codes.Unauthenticated, "invalid password")
	}

	token, err := generateJWT(req.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}

	return &pb.LoginUserResponse{
		Token: token,
	}, nil
}

func UnaryAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	fmt.Printf("Handling method: %s\n", info.FullMethod)

	if info.FullMethod == "/bookrental.BookRentalService/RegisterUser" || info.FullMethod == "/bookrental.BookRentalService/LoginUser" {
		return handler(ctx, req)
	}

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

func generateJWT(username string) (string, error) {
	secretKey := []byte("12345")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
	})

	return token.SignedString(secretKey)
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

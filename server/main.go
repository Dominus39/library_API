package main

import (
	"context"
	"fmt"
	"gc2-yugo/config"
	"gc2-yugo/entity"
	"gc2-yugo/pb"
	"log"
	"net"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
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
	booksCollection *mongo.Collection
}

type contextKey string

const usernameKey contextKey = "username"

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

func (s *BookRentalServiceServer) AddBook(ctx context.Context, req *pb.AddBookRequest) (*pb.BookResponse, error) {
	publishedDate, err := time.Parse("2006-01-02", req.PublishedDate)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid published_date format: %v", err)
	}

	newBook := entity.Book{
		ID:            primitive.NewObjectID(),
		Title:         req.Title,
		Author:        req.Author,
		PublishedDate: publishedDate,
		Status:        "Available",
	}

	_, err = s.booksCollection.InsertOne(ctx, newBook)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add book: %v", err)
	}

	return &pb.BookResponse{
		Message: "book succesfully added",
	}, nil
}

func (s *BookRentalServiceServer) RemoveBook(ctx context.Context, req *pb.RemoveBookRequest) (*pb.BookResponse, error) {
	bookID, err := primitive.ObjectIDFromHex(req.BookId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid book ID format: %v", err)
	}

	result, err := s.booksCollection.DeleteOne(ctx, bson.M{"_id": bookID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete book: %v", err)
	}

	if result.DeletedCount == 0 {
		return nil, status.Errorf(codes.NotFound, "Book not found")
	}

	return &pb.BookResponse{
		Message: "Book successfully removed",
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
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized: No metadata found")
	}

	tokenList := md["authorization"]
	if len(tokenList) == 0 {
		fmt.Println("Invalid or missing token")
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized: invalid or missing token")
	}

	token := strings.TrimPrefix(tokenList[0], "Bearer ")

	claims, err := validateJWT(token)
	if err != nil {
		fmt.Printf("Token validation failed: %v\n", err)
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized: %v", err)
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized: Invalid token claims")
	}

	ctx = context.WithValue(ctx, usernameKey, username)

	fmt.Println("Token validated successfully")
	return ctx, nil
}

func generateJWT(username string) (string, error) {
	secretKey := []byte("12345")

	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secretKey)
}

func validateJWT(tokenString string) (jwt.MapClaims, error) {
	secretKey := []byte("12345")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func main() {
	ctx := context.Background()

	usersCollection, err := config.ConnectionDatabaseUsers(ctx)
	if err != nil {
		log.Fatalf("failed to connect users database: %v", err)
	}

	booksCollection, err := config.ConnectionDatabaseBooks(ctx)
	if err != nil {
		log.Fatalf("failed to connect books database: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(UnaryAuthInterceptor),
	)
	bookRentalService := &BookRentalServiceServer{
		usersCollection: usersCollection,
		booksCollection: booksCollection,
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

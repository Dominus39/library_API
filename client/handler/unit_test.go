package handler

import (
	"context"
	"gc2-yugo/entity"
	"gc2-yugo/pb"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// Mock gRPC server implementation
type mockBookRentalServiceServer struct {
	pb.UnimplementedBookRentalServiceServer
}

var mockUser = entity.User{
	Username: "Peter Parker",
	Password: "klewear123", // Correct password for testing
}

func (s *mockBookRentalServiceServer) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	log.Println("Mock LoginUser called with:", req.Username)

	if req.Username == "validuser" && req.Password == "validpassword" {
		return &pb.LoginUserResponse{
			Token: "fake-token",
		}, nil
	}
	return nil, status.Errorf(codes.InvalidArgument, "invalid credentials")
}

func (s *mockBookRentalServiceServer) AddBook(ctx context.Context, req *pb.AddBookRequest) (*pb.BookResponse, error) {
	// Extract token from metadata to simulate token validation
	md, _ := metadata.FromIncomingContext(ctx)
	token := md.Get("authorization")
	if len(token) == 0 || token[0] != "valid-token" {
		return nil, status.Errorf(codes.Internal, "invalid or missing token")
	}

	// Simulate successful book addition
	return &pb.BookResponse{
		Message: "Success",
	}, nil
}

func (s *mockBookRentalServiceServer) RemoveBook(ctx context.Context, req *pb.RemoveBookRequest) (*pb.BookResponse, error) {
	// Extract token from metadata to simulate token validation
	md, _ := metadata.FromIncomingContext(ctx)
	token := md.Get("authorization")
	if len(token) == 0 || token[0] != "valid-token" {
		return nil, status.Errorf(codes.Internal, "invalid or missing token")
	}

	// Simulate successful book removal
	if req.BookId == "valid-book-id" {
		return &pb.BookResponse{
			Message: "Success",
		}, nil
	}

	// Simulate book not found
	return nil, status.Errorf(codes.NotFound, "book not found")
}

// Create a mock gRPC server
func setupMockGRPCServer() (*bufconn.Listener, *grpc.ClientConn) {
	listener := bufconn.Listen(1024 * 1024) // Create an in-memory listener for testing
	server := grpc.NewServer()

	// Register mock gRPC service
	pb.RegisterBookRentalServiceServer(server, &mockBookRentalServiceServer{})

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Dial the mock server
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to mock gRPC server: %v", err)
	}

	return listener, conn
}

func setupMockRemoveBookGRPCServer() (*bufconn.Listener, *grpc.ClientConn) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	pb.RegisterBookRentalServiceServer(server, &mockBookRentalServiceServer{})

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatalf("Failed to start mock gRPC server: %v", err)
		}
	}()

	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to mock gRPC server: %v", err)
	}

	return listener, conn
}

// Test LoginUser handler with a mock gRPC server
func TestLoginUser(t *testing.T) {
	// Set up the mock gRPC server and client connection
	listener, conn := setupMockGRPCServer()
	defer listener.Close()
	defer conn.Close()

	// Create Echo instance
	e := echo.New()

	// Create HTTP request and recorder
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create a mock gRPC client for this test context
	client := pb.NewBookRentalServiceClient(conn)

	// Pass the mock client into the handler context
	c.Set("grpcClient", client)

	// Simulate request body
	loginRequest := pb.LoginUserRequest{
		Username: "Peter Parker",
		Password: "klewear123",
	}
	c.Set("request", loginRequest)

	// Call the handler
	if err := LoginUser(c); err != nil {
		t.Fatalf("Failed to process request: %v", err)
	}

	// Assert that the status code is 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Assert that the response contains the expected token
	assert.Contains(t, rec.Body.String(), "fake-token")
}

func TestAddBook(t *testing.T) {
	// Set up mock gRPC server and connection
	listener, conn := setupMockGRPCServer()
	defer listener.Close()
	defer conn.Close()

	// Create Echo instance
	e := echo.New()

	// Create HTTP request and recorder
	req := httptest.NewRequest(http.MethodPost, "/book/add", nil)
	req.Header.Set("Authorization", "valid-token") // Set valid token
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock request body
	addBookRequest := pb.AddBookRequest{
		Title:  "Test Book",
		Author: "Test Author",
	}
	c.Set("request", addBookRequest)

	// Replace the gRPC client with the mock connection
	client := pb.NewBookRentalServiceClient(conn)
	c.Set("grpcClient", client)

	// Call the AddBook handler
	if err := AddBook(c); err != nil {
		t.Fatalf("Handler error: %v", err)
	}

	// Assert the status code
	assert.Equal(t, http.StatusOK, rec.Code)

	// Assert the response body contains success
	assert.Contains(t, rec.Body.String(), `"success":true`)
}

func TestRemoveBook(t *testing.T) {
	// Set up mock gRPC server and connection
	listener, conn := setupMockRemoveBookGRPCServer()
	defer listener.Close()
	defer conn.Close()

	// Create Echo instance
	e := echo.New()

	// Create HTTP request and recorder
	req := httptest.NewRequest(http.MethodDelete, "/book/remove/valid-book-id", nil)
	req.Header.Set("Authorization", "valid-token") // Set valid token
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("valid-book-id")

	// Replace the gRPC client with the mock connection
	client := pb.NewBookRentalServiceClient(conn)
	c.Set("grpcClient", client)

	// Call the RemoveBook handler
	if err := RemoveBook(c); err != nil {
		t.Fatalf("Handler error: %v", err)
	}

	// Assert the status code
	assert.Equal(t, http.StatusOK, rec.Code)

	// Assert the response body contains success
	assert.Contains(t, rec.Body.String(), `"success":true`)
}

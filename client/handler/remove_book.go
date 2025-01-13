package handler

import (
	"context"
	"gc2-yugo/pb"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// RemoveBook godoc
// @Summary Remove a book from the library
// @Description Allows an admin or authorized user to remove a book from the library
// @Tags books
// @Accept json
// @Produce json
// @Param id path string true "Book ID" example("60c72b2f9e15b92bbcf68f2b")
// @Success 200 {object} pb.RemoveBookResponse "Successfully removed the book"
// @Failure 400 {object} ErrorResponse "Invalid book ID format"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /book/remove/{id} [delete]
func RemoveBook(c echo.Context) error {

	bookID := c.Param("id")

	if _, err := primitive.ObjectIDFromHex(bookID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid book ID format")
	}

	req := &pb.RemoveBookRequest{
		BookId: bookID,
	}

	grpcConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer grpcConn.Close()

	client := pb.NewBookRentalServiceClient(grpcConn)

	token := c.Request().Header.Get("Authorization")
	if token == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
	}

	md := metadata.Pairs("authorization", token)
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	resp, err := client.RemoveBook(ctx, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

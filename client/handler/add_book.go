package handler

import (
	"context"
	"gc2-yugo/pb"
	"net/http"

	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AddBook godoc
// @Summary Add a new book
// @Description Adds a new book to the library system
// @Tags books
// @Accept json
// @Produce json
// @Param request body pb.AddBookRequest true "AddBookRequest"
// @Param Authorization header string true "Bearer <JWT Token>"
// @Success 200 {object} pb.AddBookResponse "Successfully added the book"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 401 {object} ErrorResponse "Unauthorized - missing or invalid token"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /book/add [post]
func AddBook(c echo.Context) error {
	req := new(pb.AddBookRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
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

	resp, err := client.AddBook(ctx, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)

}

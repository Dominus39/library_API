package handler

import (
	"context"
	"gc2-yugo/pb"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
)

// RegisterUser godoc
// @Summary User registration
// @Description Allows a new user to register for an account
// @Tags users
// @Accept json
// @Produce json
// @Param register_user_request body pb.RegisterUserRequest true "Register user request"
// @Success 200 {object} map[string]interface{} "Successfully registered"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /register [post]
func RegisterUser(c echo.Context) error {
	req := new(pb.RegisterUserRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	grpcConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer grpcConn.Close()

	client := pb.NewBookRentalServiceClient(grpcConn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	resp, err := client.RegisterUser(ctx, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": resp.Message,
		"user_id": resp.UserId,
	})

}

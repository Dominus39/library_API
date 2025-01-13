package handler

import (
	"context"
	"gc2-yugo/pb"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
)

// LoginUser godoc
// @Summary User login
// @Description Allows a user to log in using their credentials
// @Tags users
// @Accept json
// @Produce json
// @Param login_user_request body pb.LoginUserRequest true "Login user request"
// @Success 200 {object} pb.LoginUserResponse "Successfully logged in"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 401 {object} ErrorResponse "Unauthorized - invalid credentials"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /login [post]
func LoginUser(c echo.Context) error {
	req := new(pb.LoginUserRequest)
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

	resp, err := client.LoginUser(ctx, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

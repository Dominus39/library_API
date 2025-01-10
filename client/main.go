package main

import (
	"gc2-yugo/client/handler"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.POST("/register", handler.RegisterUser)

	e.Logger.Fatal(e.Start(":8080"))
}

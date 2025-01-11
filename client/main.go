package main

import (
	"gc2-yugo/client/handler"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.POST("/register", handler.RegisterUser)
	e.POST("/login", handler.LoginUser)
	e.POST("/book/add", handler.AddBook)
	e.DELETE("/book/remove/:id", handler.RemoveBook)

	e.Logger.Fatal(e.Start(":8080"))
}

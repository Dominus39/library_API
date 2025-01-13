package main

import (
	"gc2-yugo/client/handler"
	"gc2-yugo/utils"

	_ "gc2-yugo/client/docs" // This will import your generated docs

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title Book Borrowing API
// @version 1.0
// @description This API allows users to manage books in a library, including registering, logging in, adding/removing books, and borrowing books.
// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com
// @license.name MIT
// @host localhost:8080
// @BasePath /
// @schemes http https
func main() {

	utils.StartSchedulerJob()

	e := echo.New()

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.POST("/register", handler.RegisterUser)
	e.POST("/login", handler.LoginUser)
	e.POST("/book/add", handler.AddBook)
	e.DELETE("/book/remove/:id", handler.RemoveBook)
	e.POST("/book/borrow/:id", handler.BorrowBook)

	e.Logger.Fatal(e.Start(":8080"))
}

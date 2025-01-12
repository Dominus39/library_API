package main

import (
	"gc2-yugo/client/handler"
	"gc2-yugo/utils"

	"github.com/labstack/echo/v4"
)

func main() {

	utils.StartSchedulerJob()

	e := echo.New()

	e.POST("/register", handler.RegisterUser)
	e.POST("/login", handler.LoginUser)
	e.POST("/book/add", handler.AddBook)
	e.DELETE("/book/remove/:id", handler.RemoveBook)
	e.POST("/book/borrow/:id", handler.BorrowBook)

	e.Logger.Fatal(e.Start(":8080"))
}

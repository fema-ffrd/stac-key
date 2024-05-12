package main

import (
	config "keyval/config"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	ADMIN  bool = true
	PUBLIC bool = false
)

func main() {
	appConfig := config.Init()
	allUsers := []string{"s3_admin"}

	// Instantiate server
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: true,
		AllowOrigins:     []string{"*"},
	}))

	e.GET("/", config.Ping(appConfig))
	e.GET("/ping", config.Ping(appConfig))
	e.GET("/validate", config.Authorize(config.Validate(appConfig), allUsers...))

	e.Logger.Fatal(e.Start(appConfig.Address()))
}

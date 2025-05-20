package main

import (
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mccune1224/strider/handler"
)

const devPort = "5173"

func getEnvPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return devPort
	}
	return port
}

func main() {
	fmt.Println("We are so online")

	app := echo.New()
	app.Static("/", "assets")

	app.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${status} ${latency_human} ${method} ${path}\n",
	}))

	h := &handler.Handler{}
	h.AttachRoutes(app)

	app.Logger.Fatal(app.Start(":" + getEnvPort()))
}

package main

import (
	"jwt-poc/app/api/routes"
	"jwt-poc/config"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	config.ConnectDB()

	app := fiber.New()
	routes.RegisterRoutes(app)

	port := os.Getenv("APP_PORT")

	_ = app.Listen(":" + port)
}

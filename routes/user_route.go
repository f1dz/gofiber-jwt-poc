package routes

import (
	"jwt-poc/handlers"
	"jwt-poc/middlewares"

	"github.com/gofiber/fiber/v2"
)

func UserRoutes(router fiber.Router) {
	user := router.Group("/user")
	user.Post("/register", handlers.CreateUserHandler)
	user.Use(middlewares.AuthMiddleware())
	user.Get("/profile", handlers.ProfileHandler)
}

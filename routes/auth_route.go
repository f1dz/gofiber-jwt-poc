package routes

import (
	"jwt-poc/handlers"

	"github.com/gofiber/fiber/v2"
)

func AuthRoute(router fiber.Router) {
	auth := router.Group("/auth")

	auth.Post("/login", handlers.LoginHandler)
	auth.Post("/refresh", handlers.RefreshTokenHandler)
}

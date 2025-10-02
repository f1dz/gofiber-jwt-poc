package routes

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(app *fiber.App) {
	api := app.Group("/api")
	AuthRoute(api)
	UserRoutes(api)
}

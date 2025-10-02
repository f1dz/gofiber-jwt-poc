package middlewares

import (
	"jwt-poc/config"
	"jwt-poc/models"
	"jwt-poc/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		apiKeyHeader := c.Get("api-key")

		// 🔹 1. Cek Authorization (Bearer JWT)
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid or malformed Authorization header",
				})
			}

			tokenString := parts[1]

			// Validate JWT token
			claims, err := utils.ValidateJWT(tokenString)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid or expired JWT",
				})
			}

			// Store user information in context
			c.Locals("userID", claims.UserID)
			c.Locals("role", claims.Role)
			c.Locals("authType", "JWT")

			return c.Next()
		}

		// 🔹 2. Cek X-API-Key
		if apiKeyHeader != "" {
			var apiKey models.ApiKey
			if err := config.DB.Where("key = ? AND is_active = ?", apiKeyHeader, true).First(&apiKey).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
						"error": "Invalid or inactive API key",
					})
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Internal server error",
				})
			}

			c.Locals("clientID", apiKey.Client)
			c.Locals("scope", apiKey.Scope)
			c.Locals("userID", apiKey.UserID)
			c.Locals("authType", "APIKey")

			return c.Next()
		}

		// 🔹 Kalau dua-duanya kosong
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing authentication (JWT or API Key)",
		})
	}
}

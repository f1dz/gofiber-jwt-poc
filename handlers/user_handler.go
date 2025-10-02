package handlers

import (
	"jwt-poc/config"
	"jwt-poc/models"
	"jwt-poc/utils"

	"github.com/gofiber/fiber/v2"
)

func CreateUserHandler(c *fiber.Ctx) error {
	type CreateUserRequest struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Role     string `json:"role" validate:"required,oneof=admin user"`
	}

	request := CreateUserRequest{}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	var dbUser models.User
	config.DB.Where("username = ?", request.Username).First(&dbUser)
	if dbUser.ID != 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username already exists",
		})
	}

	hashedPassword, err := utils.HashPassword(request.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	newUser := models.User{
		Username:     request.Username,
		PasswordHash: hashedPassword,
		Email:        request.Email,
		Role:         request.Role,
	}

	config.DB.Create(&newUser)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"user":    newUser,
	})
}

func ProfileHandler(c *fiber.Ctx) error {
	authType := c.Locals("authType").(string)
	if authType == "JWT" {
		userID := c.Locals("userID").(uint)
		role := c.Locals("role").(string)
		return c.JSON(fiber.Map{
			"user_id":   userID,
			"role":      role,
			"access_by": authType,
		})
	} else if authType == "APIKey" {
		clientID := c.Locals("clientID").(string)
		role := c.Locals("scope").(string)
		return c.JSON(fiber.Map{
			"client_id": clientID,
			"role":      role,
			"access_by": authType,
			"user_id":   c.Locals("userID"),
		})
	}

	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error": "Unauthorized access",
	})
}

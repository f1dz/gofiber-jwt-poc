package config

import (
	"fmt"
	"jwt-poc/models"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dbName := "gofiber_auth.db"
	var err error

	DB, err = gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database", err)
	}

	fmt.Println("Database connected successfully")

	err = DB.AutoMigrate(&models.User{}, &models.RefreshToken{}, &models.ApiKey{})

	if err != nil {
		log.Fatal("failed to migrate database")
	}

	fmt.Println("Database migrated successfully")
}

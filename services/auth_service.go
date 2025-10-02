package services

import (
	"jwt-poc/config"
	"jwt-poc/models"
	"jwt-poc/utils"
	"time"

	"github.com/google/uuid"
)

func GenerateAuthToken(user models.User) (accessToken string, refreshToken string, err error) {
	accessToken, err = utils.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return "", "", err
	}

	refreshToken = uuid.New().String()
	expiry := time.Now().Add(30 * 24 * time.Hour)

	refreshTokenModel := models.RefreshToken{
		UserID:     user.ID,
		Token:      refreshToken,
		ExpiryDate: expiry,
	}

	if err := config.DB.Create(&refreshTokenModel).Error; err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func RefreshAndRevokeToken(oldRefreshToken string) (accessToken string, newRefreshToken string, err error) {
	var oldToken models.RefreshToken
	if err := config.DB.Where("token = ? AND expiry_date > ?", oldRefreshToken, time.Now()).First(&oldToken).Error; err != nil {
		return "", "", err
	}

	var user models.User
	if err := config.DB.First(&user, oldToken.UserID).Error; err != nil {
		return "", "", err
	}

	config.DB.Delete(&oldToken)

	accessToken, newRefreshToken, err = GenerateAuthToken(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

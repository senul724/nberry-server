package services

import (
	"IAM-server/internal/connections"
	"IAM-server/internal/models"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GenerateAppSecret generates a cryptographically secure random API secret key.
// Format: nb_sec_<64 hex characters>
func GenerateAppSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure secret: %w", err)
	}
	return "nb_sec_" + hex.EncodeToString(bytes), nil
}

// HashAppSecret calculates the SHA-256 checksum of the secret key.
func HashAppSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hash[:])
}

// CreateApp creates a new application owned by the user, generating and hashing an API secret.
// Returns the created app and the raw secret key (only shown once).
func CreateApp(ctx context.Context, userID string, name string, description string, logo *string, unicastCallbackURL *string) (*models.App, string, error) {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, "", fmt.Errorf("invalid user ID: %w", err)
	}

	rawSecret, err := GenerateAppSecret()
	if err != nil {
		return nil, "", err
	}

	secretHash := HashAppSecret(rawSecret)

	app := models.App{
		UserID:             uUUID,
		Name:               name,
		Description:        description,
		Logo:               logo,
		SecretHash:         secretHash,
		UnicastCallbackURL: unicastCallbackURL,
	}

	if err := connections.DB.WithContext(ctx).Create(&app).Error; err != nil {
		return nil, "", fmt.Errorf("failed to create app: %w", err)
	}

	return &app, rawSecret, nil
}

// DeleteApp deletes an app by ID owned by the user.
func DeleteApp(ctx context.Context, userID string, appID string) error {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	aUUID, err := uuid.Parse(appID)
	if err != nil {
		return fmt.Errorf("invalid app ID: %w", err)
	}

	var app models.App
	err = connections.DB.WithContext(ctx).
		Where("id = ? AND user_id = ?", aUUID, uUUID).
		First(&app).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("app not found")
		}
		return fmt.Errorf("failed to find app: %w", err)
	}

	if err := connections.DB.WithContext(ctx).Delete(&app).Error; err != nil {
		return fmt.Errorf("failed to delete app: %w", err)
	}

	return nil
}

// GetUserApps returns all apps owned by a user.
func GetUserApps(ctx context.Context, userID string) ([]models.App, error) {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var apps []models.App
	err = connections.DB.WithContext(ctx).
		Where("user_id = ?", uUUID).
		Order("created_at desc").
		Find(&apps).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list apps: %w", err)
	}

	return apps, nil
}

// GetAppByID retrieves a single app by ID owned by the user.
func GetAppByID(ctx context.Context, userID string, appID string) (*models.App, error) {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	aUUID, err := uuid.Parse(appID)
	if err != nil {
		return nil, fmt.Errorf("invalid app ID: %w", err)
	}

	var app models.App
	err = connections.DB.WithContext(ctx).
		Where("id = ? AND user_id = ?", aUUID, uUUID).
		First(&app).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("app not found")
		}
		return nil, fmt.Errorf("failed to find app: %w", err)
	}

	return &app, nil
}

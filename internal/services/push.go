package services

import (
	"IAM-server/internal/connections"
	"IAM-server/internal/models"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PushBroadcast creates a new broadcast memo in the database for the given app.
func PushBroadcast(ctx context.Context, appID uuid.UUID, title string, content string, image *string) (*models.BroadcastMemo, error) {
	memo := models.BroadcastMemo{
		AppID:   appID,
		Title:   title,
		Content: content,
		Image:   image,
	}

	if err := connections.DB.WithContext(ctx).Create(&memo).Error; err != nil {
		return nil, fmt.Errorf("failed to save broadcast memo: %w", err)
	}

	return &memo, nil
}

// PushUnicast resolves the target user using either unicastKey or userIDStr,
// verifies the unicast subscription exists for this app, and creates a direct memo in the database.
func PushUnicast(ctx context.Context, appID uuid.UUID, unicastKey string, userIDStr string, title string, content string, image *string) (*models.DirectMemo, error) {
	unicastKey = strings.TrimSpace(unicastKey)
	userIDStr = strings.TrimSpace(userIDStr)

	if unicastKey == "" || userIDStr == "" {
		return nil, errors.New("both unicast_key and user_id are required")
	}

	targetUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	// Look up subscription for this app and user
	var sub models.UnicastSub
	err = connections.DB.WithContext(ctx).
		Where("app_id = ? AND user_id = ?", appID, targetUserID).
		First(&sub).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("unicast subscription not found for the provided user")
		}
		return nil, fmt.Errorf("failed to query unicast subscription: %w", err)
	}

	// Validate the provided unicast key hash
	keyHash := HashKey(unicastKey)
	if sub.UnicastKeyHash != keyHash {
		return nil, errors.New("invalid unicast_key")
	}

	memo := models.DirectMemo{
		AppID:   appID,
		UserID:  targetUserID,
		Title:   title,
		Content: content,
		Image:   image,
	}

	if err := connections.DB.WithContext(ctx).Create(&memo).Error; err != nil {
		return nil, fmt.Errorf("failed to save direct memo: %w", err)
	}

	return &memo, nil
}

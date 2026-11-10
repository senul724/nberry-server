package services

import (
	"IAM-server/internal/connections"
	"IAM-server/internal/models"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerSubscriptionsResponse struct {
	Broadcast []models.BroadcastSub `json:"broadcast"`
	Unicast   []models.UnicastSub   `json:"unicast"`
}

// UpdateNotifyToken updates the device push notification token (FCM/APNs) for the specified user.
func UpdateNotifyToken(ctx context.Context, userID string, notifyToken *string) error {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	result := connections.DB.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", uUUID).
		Update("notify_token", notifyToken)

	if result.Error != nil {
		return fmt.Errorf("failed to update notify token: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// GetCustomerBroadcastMessages retrieves broadcast memos from apps the user has subscribed to.
// Can be optionally filtered by appID.
func GetCustomerBroadcastMessages(ctx context.Context, userID string, appID string, limit int, offset int) ([]models.BroadcastMemo, error) {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	subQuery := connections.DB.Table("broadcast_subs").
		Select("app_id").
		Where("user_id = ?", uUUID)

	query := connections.DB.WithContext(ctx).
		Model(&models.BroadcastMemo{}).
		Where("app_id IN (?)", subQuery)

	if appID != "" {
		aUUID, err := uuid.Parse(appID)
		if err != nil {
			return nil, fmt.Errorf("invalid app ID: %w", err)
		}
		query = query.Where("app_id = ?", aUUID)
	}

	var memos []models.BroadcastMemo
	if err := query.
		Preload("App", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, logo, description, created_at, updated_at")
		}).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&memos).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve broadcast memos: %w", err)
	}

	if memos == nil {
		memos = []models.BroadcastMemo{}
	}

	return memos, nil
}

// GetCustomerDirectMessages retrieves direct memos addressed to the user.
// Can be optionally filtered by appID.
func GetCustomerDirectMessages(ctx context.Context, userID string, appID string, limit int, offset int) ([]models.DirectMemo, error) {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := connections.DB.WithContext(ctx).
		Model(&models.DirectMemo{}).
		Where("user_id = ?", uUUID)

	if appID != "" {
		aUUID, err := uuid.Parse(appID)
		if err != nil {
			return nil, fmt.Errorf("invalid app ID: %w", err)
		}
		query = query.Where("app_id = ?", aUUID)
	}

	var memos []models.DirectMemo
	if err := query.
		Preload("App", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, logo, description, created_at, updated_at")
		}).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&memos).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve direct memos: %w", err)
	}

	if memos == nil {
		memos = []models.DirectMemo{}
	}

	return memos, nil
}

// GetCustomerBroadcastSubscriptions retrieves all broadcast subscriptions for a user with vendor app details.
func GetCustomerBroadcastSubscriptions(ctx context.Context, userID string) ([]models.BroadcastSub, error) {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var subs []models.BroadcastSub
	if err := connections.DB.WithContext(ctx).
		Where("user_id = ?", uUUID).
		Preload("App", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, logo, description, created_at, updated_at")
		}).
		Order("created_at DESC").
		Find(&subs).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve broadcast subscriptions: %w", err)
	}

	if subs == nil {
		subs = []models.BroadcastSub{}
	}

	return subs, nil
}

// GetCustomerUnicastSubscriptions retrieves all unicast subscriptions for a user with vendor app details.
func GetCustomerUnicastSubscriptions(ctx context.Context, userID string) ([]models.UnicastSub, error) {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var subs []models.UnicastSub
	if err := connections.DB.WithContext(ctx).
		Where("user_id = ?", uUUID).
		Preload("App", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, logo, description, created_at, updated_at")
		}).
		Order("created_at DESC").
		Find(&subs).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve unicast subscriptions: %w", err)
	}

	if subs == nil {
		subs = []models.UnicastSub{}
	}

	return subs, nil
}

// GetCustomerSubscriptions retrieves both broadcast and unicast subscriptions for a user.
func GetCustomerSubscriptions(ctx context.Context, userID string) (*CustomerSubscriptionsResponse, error) {
	broadcastSubs, err := GetCustomerBroadcastSubscriptions(ctx, userID)
	if err != nil {
		return nil, err
	}

	unicastSubs, err := GetCustomerUnicastSubscriptions(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &CustomerSubscriptionsResponse{
		Broadcast: broadcastSubs,
		Unicast:   unicastSubs,
	}, nil
}

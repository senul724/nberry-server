package services

import (
	"IAM-server/internal/connections"
	"IAM-server/internal/models"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GenerateUnicastKey generates a cryptographically secure random key for unicast notifications.
// Format: nb_uni_<64 hex characters>
func GenerateUnicastKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}
	return "nb_uni_" + hex.EncodeToString(b), nil
}

// HashKey computes the SHA-256 hash of a string and returns it in hex format.
func HashKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

// SubscribeBroadcast subscribes a user to an app's broadcast channel.
func SubscribeBroadcast(ctx context.Context, userID string, appID string) (*models.BroadcastSub, error) {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	aUUID, err := uuid.Parse(appID)
	if err != nil {
		return nil, fmt.Errorf("invalid app ID: %w", err)
	}

	// Verify app exists
	var app models.App
	if err := connections.DB.WithContext(ctx).First(&app, "id = ?", aUUID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("app not found")
		}
		return nil, fmt.Errorf("failed to verify app: %w", err)
	}

	var existing models.BroadcastSub
	err = connections.DB.WithContext(ctx).
		Where("app_id = ? AND user_id = ?", aUUID, uUUID).
		First(&existing).Error

	if err == nil {
		// Already subscribed
		return &existing, nil
	}

	sub := models.BroadcastSub{
		AppID:  aUUID,
		UserID: uUUID,
	}

	if err := connections.DB.WithContext(ctx).Create(&sub).Error; err != nil {
		return nil, fmt.Errorf("failed to subscribe to broadcast: %w", err)
	}

	return &sub, nil
}

// UnsubscribeBroadcast removes a user's subscription to an app's broadcast channel.
func UnsubscribeBroadcast(ctx context.Context, userID string, appID string) error {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	aUUID, err := uuid.Parse(appID)
	if err != nil {
		return fmt.Errorf("invalid app ID: %w", err)
	}

	result := connections.DB.WithContext(ctx).
		Where("app_id = ? AND user_id = ?", aUUID, uUUID).
		Delete(&models.BroadcastSub{})

	if result.Error != nil {
		return fmt.Errorf("failed to unsubscribe: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("subscription not found")
	}

	return nil
}

type UnicastCallbackPayload struct {
	UnicastKey   string `json:"unicast_key"`
	VendorUserID string `json:"vendor_user_id"`
}

// dispatchUnicastCallback posts the unicast key and vendor user ID to the app's callback URL.
func dispatchUnicastCallback(callbackURL, unicastKey, vendorUserID string) error {
	payload := UnicastCallbackPayload{
		UnicastKey:   unicastKey,
		VendorUserID: vendorUserID,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to serialize callback payload: %w", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, callbackURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create callback request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("callback request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("callback returned non-2xx status code: %d", resp.StatusCode)
	}

	return nil
}

// SubscribeUnicast subscribes a user to an app's unicast channel, hashes the secret key,
// and invokes the vendor's callback URL with the raw unicast key and vendorUserID.
func SubscribeUnicast(ctx context.Context, userID string, appID string, vendorUserID string) (*models.UnicastSub, string, error) {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, "", fmt.Errorf("invalid user ID: %w", err)
	}

	aUUID, err := uuid.Parse(appID)
	if err != nil {
		return nil, "", fmt.Errorf("invalid app ID: %w", err)
	}

	var app models.App
	if err := connections.DB.WithContext(ctx).First(&app, "id = ?", aUUID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", errors.New("app not found")
		}
		return nil, "", fmt.Errorf("failed to verify app: %w", err)
	}

	rawKey, err := GenerateUnicastKey()
	if err != nil {
		return nil, "", err
	}

	keyHash := HashKey(rawKey)

	var sub models.UnicastSub
	err = connections.DB.WithContext(ctx).
		Where("app_id = ? AND user_id = ?", aUUID, uUUID).
		First(&sub).Error

	if err == nil {
		// Update existing unicast key hash
		sub.UnicastKeyHash = keyHash
		if err := connections.DB.WithContext(ctx).Save(&sub).Error; err != nil {
			return nil, "", fmt.Errorf("failed to update unicast subscription: %w", err)
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new unicast subscription
		sub = models.UnicastSub{
			AppID:          aUUID,
			UserID:         uUUID,
			UnicastKeyHash: keyHash,
		}
		if err := connections.DB.WithContext(ctx).Create(&sub).Error; err != nil {
			return nil, "", fmt.Errorf("failed to create unicast subscription: %w", err)
		}
	} else {
		return nil, "", fmt.Errorf("database query error: %w", err)
	}

	// Trigger callback to app's unicast_callback_url if configured
	if app.UnicastCallbackURL != nil && *app.UnicastCallbackURL != "" {
		if err := dispatchUnicastCallback(*app.UnicastCallbackURL, rawKey, vendorUserID); err != nil {
			log.Printf("[UnicastCallback] warning: callback to %s failed: %v", *app.UnicastCallbackURL, err)
			// Return warning along with the sub or continue; we log it so subscription is not lost
		}
	}

	return &sub, rawKey, nil
}

// UnsubscribeUnicast removes a user's unicast subscription for an app.
func UnsubscribeUnicast(ctx context.Context, userID string, appID string) error {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	aUUID, err := uuid.Parse(appID)
	if err != nil {
		return fmt.Errorf("invalid app ID: %w", err)
	}

	result := connections.DB.WithContext(ctx).
		Where("app_id = ? AND user_id = ?", aUUID, uUUID).
		Delete(&models.UnicastSub{})

	if result.Error != nil {
		return fmt.Errorf("failed to unsubscribe from unicast: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("subscription not found")
	}

	return nil
}

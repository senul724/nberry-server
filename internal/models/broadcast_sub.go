package models

import (
	"time"

	"github.com/google/uuid"
)

type BroadcastSub struct {
	AppID     uuid.UUID `gorm:"type:uuid;primaryKey"                             json:"app_id"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"                             json:"user_id"`
	CreatedAt time.Time `gorm:"autoCreateTime;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Associations with cascade on delete
	App  App  `gorm:"foreignKey:AppID;constraint:OnDelete:CASCADE"  json:"app,omitempty"`
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

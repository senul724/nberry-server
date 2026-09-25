package models

import (
	"github.com/google/uuid"
)

type UnicastSub struct {
	Base
	AppID          uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_app_user_unicast" json:"app_id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_app_user_unicast" json:"user_id"`
	UnicastKeyHash string    `gorm:"type:text;not null"                                        json:"-"`

	// Associations with cascade on delete
	App  App  `gorm:"foreignKey:AppID;constraint:OnDelete:CASCADE"  json:"app,omitempty"`
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

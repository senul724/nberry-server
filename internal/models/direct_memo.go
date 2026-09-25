package models

import (
	"github.com/google/uuid"
)

type DirectMemo struct {
	Base
	AppID   uuid.UUID `gorm:"type:uuid;not null;index" json:"app_id"`
	UserID  uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Title   string    `gorm:"type:text;not null"       json:"title"`
	Content string    `gorm:"type:text;not null"       json:"content"`
	Image   *string   `gorm:"type:text;default:null"   json:"image,omitempty"`

	// Associations with cascade on delete
	App  App  `gorm:"foreignKey:AppID;constraint:OnDelete:CASCADE"  json:"app,omitempty"`
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

package models

import (
	"github.com/google/uuid"
)

type BroadcastMemo struct {
	Base
	AppID   uuid.UUID `gorm:"type:uuid;not null;index" json:"app_id"`
	Title   string    `gorm:"type:text;not null"       json:"title"`
	Content string    `gorm:"type:text;not null"       json:"content"`
	Image   *string   `gorm:"type:text;default:null"   json:"image,omitempty"`

	// Association with cascade on delete
	App App `gorm:"foreignKey:AppID;constraint:OnDelete:CASCADE" json:"app,omitempty"`
}

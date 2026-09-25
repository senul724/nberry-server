package models

import (
	"github.com/google/uuid"
)

type App struct {
	Base
	UserID             uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Name               string    `gorm:"type:text;not null"       json:"name"`
	Logo               *string   `gorm:"type:text;default:null"   json:"logo,omitempty"`
	Description        string    `gorm:"type:text"                json:"description"`
	SecretHash         string    `gorm:"type:text;not null"       json:"-"`
	UnicastCallbackURL *string   `gorm:"type:text;default:null"   json:"unicast_callback_url,omitempty"`

	// Associations
	User           User            `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	BroadcastSubs  []BroadcastSub  `gorm:"foreignKey:AppID;constraint:OnDelete:CASCADE"  json:"broadcast_subs,omitempty"`
	BroadcastMemos []BroadcastMemo `gorm:"foreignKey:AppID;constraint:OnDelete:CASCADE"  json:"broadcast_memos,omitempty"`
	UnicastSubs    []UnicastSub    `gorm:"foreignKey:AppID;constraint:OnDelete:CASCADE"  json:"unicast_subs,omitempty"`
	DirectMemos    []DirectMemo    `gorm:"foreignKey:AppID;constraint:OnDelete:CASCADE"  json:"direct_memos,omitempty"`
}

package models

type User struct {
	Base
	Name        string  `gorm:"not null"                     json:"name"`
	Email       string  `gorm:"uniqueIndex;not null"         json:"email"`
	Password    *string `gorm:"default:null"                 json:"-"`
	PhotoURL    *string `gorm:"default:null"                 json:"photo_url,omitempty"`
	NotifyToken *string `gorm:"default:null"                 json:"notify_token,omitempty"`

	// Associations
	Apps          []App          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"apps,omitempty"`
	BroadcastSubs []BroadcastSub `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"broadcast_subs,omitempty"`
	UnicastSubs   []UnicastSub   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"unicast_subs,omitempty"`
	DirectMemos   []DirectMemo   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"direct_memos,omitempty"`
	Sessions      []Session      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"sessions,omitempty"`
}

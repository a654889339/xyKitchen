package models

import "time"

type Message struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	UserID    int       `gorm:"column:userId;not null;index" json:"userId"`
	Sender    string    `gorm:"type:enum('user','admin');not null;default:user" json:"sender"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Type      string    `gorm:"size:10;default:text" json:"type"`
	Read      bool      `gorm:"column:read" json:"read"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Message) TableName() string { return "messages" }

type OutletMessage struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	UserID    int       `gorm:"column:userId;not null;index" json:"userId"`
	Sender    string    `gorm:"type:enum('user','admin');not null;default:user" json:"sender"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Type      string    `gorm:"size:10;default:text" json:"type"`
	Read      bool      `gorm:"column:read" json:"read"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (OutletMessage) TableName() string { return "outlet_messages" }

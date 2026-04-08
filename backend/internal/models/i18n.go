package models

import "time"

type I18nText struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"size:100;not null;uniqueIndex" json:"key"`
	Zh        string    `gorm:"size:500;not null" json:"zh"`
	En        string    `gorm:"size:500;not null" json:"en"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (I18nText) TableName() string { return "i18n_texts" }

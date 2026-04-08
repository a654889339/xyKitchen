package models

import "time"

// 与 Sequelize 一致：复合唯一 (page_key, visit_date)，无主键 id
type PageVisitDaily struct {
	PageKey   string    `gorm:"column:page_key;size:255;primaryKey" json:"pageKey"`
	VisitDate time.Time `gorm:"column:visit_date;type:date;primaryKey" json:"visitDate"`
	Count     uint      `gorm:"not null;default:0" json:"count"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (PageVisitDaily) TableName() string { return "page_visit_daily" }

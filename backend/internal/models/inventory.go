package models

import "time"

type InventoryCategory struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	SortOrder int       `gorm:"column:sortOrder" json:"sortOrder"`
	Status    string    `gorm:"type:enum('active','inactive');default:active" json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Products  []InventoryProduct `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}

func (InventoryCategory) TableName() string { return "inventory_categories" }

type InventoryProduct struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	CategoryID   int       `gorm:"column:categoryId;not null;index" json:"categoryId"`
	Name         string    `gorm:"size:200;not null" json:"name"`
	SerialNumber string    `gorm:"column:serialNumber;size:100;not null;uniqueIndex:serialNumber" json:"serialNumber"`
	GuideSlug    string    `gorm:"column:guideSlug;size:200" json:"guideSlug"`
	SortOrder    int       `gorm:"column:sortOrder" json:"sortOrder"`
	Status       string    `gorm:"type:enum('active','inactive');default:active" json:"status"`
	Tags         string    `gorm:"size:200" json:"tags"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Category     *InventoryCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (InventoryProduct) TableName() string { return "inventory_products" }

type UserProduct struct {
	ID         int       `gorm:"primaryKey" json:"id"`
	UserID     int       `gorm:"column:userId;not null;index" json:"userId"`
	ProductKey string    `gorm:"column:productKey;size:100;not null;uniqueIndex:productKey" json:"productKey"`
	CreatedAt  time.Time `gorm:"column:createdAt" json:"createdAt"`
}

func (UserProduct) TableName() string { return "user_products" }

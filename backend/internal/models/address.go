package models

import "time"

type Address struct {
	ID             int       `gorm:"primaryKey" json:"id"`
	UserID         int       `gorm:"column:userId;not null;index" json:"userId"`
	ContactName    string    `gorm:"column:contactName;size:50;not null" json:"contactName"`
	ContactPhone   string    `gorm:"column:contactPhone;size:20;not null" json:"contactPhone"`
	Country        string    `gorm:"size:50" json:"country"`
	CustomCountry  string    `gorm:"column:customCountry;size:100" json:"customCountry"`
	Province       string    `gorm:"size:50" json:"province"`
	City           string    `gorm:"size:50" json:"city"`
	District       string    `gorm:"size:50" json:"district"`
	DetailAddress  string    `gorm:"column:detailAddress;size:500" json:"detailAddress"`
	IsDefault      bool      `gorm:"column:isDefault" json:"isDefault"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (Address) TableName() string { return "addresses" }

type OutletAddress struct {
	ID            int       `gorm:"primaryKey" json:"id"`
	UserID        int       `gorm:"column:userId;not null" json:"userId"`
	ContactName   string    `gorm:"column:contactName;size:50;not null" json:"contactName"`
	ContactPhone  string    `gorm:"column:contactPhone;size:20;not null" json:"contactPhone"`
	Country       string    `gorm:"size:50" json:"country"`
	CustomCountry string    `gorm:"column:customCountry;size:100" json:"customCountry"`
	Province      string    `gorm:"size:50" json:"province"`
	City          string    `gorm:"size:50" json:"city"`
	District      string    `gorm:"size:50" json:"district"`
	DetailAddress string    `gorm:"column:detailAddress;size:500" json:"detailAddress"`
	IsDefault     bool      `gorm:"column:isDefault" json:"isDefault"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (OutletAddress) TableName() string { return "outlet_addresses" }

package models

import "time"

// BookingConfig 单例预订全局配置（仅 xyKitchen 使用）
type BookingConfig struct {
	ID               int       `gorm:"primaryKey" json:"id"`
	HomepageBgURL    string    `gorm:"column:homepageBgUrl;size:1024" json:"homepageBgUrl"`
	NoticeTitle      string    `gorm:"column:noticeTitle;size:300" json:"noticeTitle"`
	NoticeBody       string    `gorm:"column:noticeBody;type:longtext" json:"noticeBody"`
	PerPersonDeposit float64   `gorm:"column:perPersonDeposit;type:decimal(10,2);default:50" json:"perPersonDeposit"`
	TimeSlotsJSON    string    `gorm:"column:timeSlotsJson;type:text" json:"timeSlotsJson"`       // ["17:00","19:00","21:00"]
	GuestOptionsJSON string    `gorm:"column:guestOptionsJson;type:text" json:"guestOptionsJson"` // [2,3,4,6]
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

func (BookingConfig) TableName() string { return "booking_configs" }

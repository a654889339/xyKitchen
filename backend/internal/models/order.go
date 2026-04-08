package models

import "time"

// Order 门店预订（xyKitchen）。对外 JSON 仅暴露预订相关字段；下列带 json:"-" 的为兼容旧表结构保留。
type Order struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	OrderNo      string    `gorm:"column:orderNo;size:32;not null;uniqueIndex:orderNo" json:"orderNo"`
	UserID       int       `gorm:"column:userId;not null;index" json:"userId"`
	BookingAt    *time.Time `gorm:"column:bookingAt;index" json:"bookingAt"`
	GuestCount   int        `gorm:"column:guestCount;not null;default:0" json:"guestCount"`
	ContactPhone string    `gorm:"column:contactPhone;size:20;not null;default:''" json:"contactPhone"`
	Price        float64   `gorm:"column:price;type:decimal(10,2);not null" json:"price"`
	Status       string    `gorm:"type:enum('pending','paid','processing','completed','cancelled');default:pending" json:"status"`
	Remark       string    `gorm:"type:text" json:"remark"`
	AdminRemark  string    `gorm:"column:adminRemark;type:text" json:"adminRemark"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	User         *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`

	ServiceID       *int       `gorm:"column:serviceId" json:"-"`
	ServiceTitle    string     `gorm:"column:serviceTitle;size:200" json:"-"`
	ServiceTitleEn  string     `gorm:"column:serviceTitleEn;size:200" json:"-"`
	ServiceIcon     string     `gorm:"column:serviceIcon;size:100" json:"-"`
	ContactName     string     `gorm:"column:contactName;size:50" json:"-"`
	Address         string     `gorm:"size:500" json:"-"`
	AppointmentTime *time.Time `gorm:"column:appointmentTime" json:"-"`
	ProductSerial   string     `gorm:"column:productSerial;size:128" json:"-"`
	GuideID         *int       `gorm:"column:guideId" json:"-"`
}

func (Order) TableName() string { return "orders" }

type OrderLog struct {
	ID         int       `gorm:"primaryKey" json:"id"`
	OrderID    int       `gorm:"column:orderId;not null;index" json:"orderId"`
	ChangeType string    `gorm:"column:changeType;type:enum('status','price','admin_remark');not null" json:"changeType"`
	OldValue   string    `gorm:"column:oldValue;size:500" json:"oldValue"`
	NewValue   string    `gorm:"column:newValue;size:500" json:"newValue"`
	Operator   string    `gorm:"size:100" json:"operator"`
	CreatedAt  time.Time `gorm:"column:createdAt" json:"createdAt"`
}

func (OrderLog) TableName() string { return "order_logs" }

type OutletOrder struct {
	ID              int        `gorm:"primaryKey" json:"id"`
	OrderNo         string     `gorm:"column:orderNo;size:32;not null;uniqueIndex:outlet_orders_orderNo" json:"orderNo"`
	UserID          int        `gorm:"column:userId;not null" json:"userId"`
	ServiceID       *int       `gorm:"column:serviceId" json:"serviceId"`
	ServiceTitle    string     `gorm:"column:serviceTitle;size:200;not null" json:"serviceTitle"`
	ServiceIcon     string     `gorm:"column:serviceIcon;size:100" json:"serviceIcon"`
	Price           float64    `gorm:"column:price;type:decimal(10,2)" json:"price"`
	Status          string     `gorm:"type:enum('pending','paid','processing','completed','cancelled');default:pending" json:"status"`
	Remark          string     `gorm:"type:text" json:"remark"`
	ContactName     string     `gorm:"column:contactName;size:50" json:"contactName"`
	ContactPhone    string     `gorm:"column:contactPhone;size:20" json:"contactPhone"`
	Address         string     `gorm:"size:500" json:"address"`
	AppointmentTime *time.Time `gorm:"column:appointmentTime" json:"appointmentTime"`
	AdminRemark     string     `gorm:"column:adminRemark;type:text" json:"adminRemark"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	User            *OutletUser `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (OutletOrder) TableName() string { return "outlet_orders" }

type OutletOrderLog struct {
	ID         int       `gorm:"primaryKey" json:"id"`
	OrderID    int       `gorm:"column:orderId;not null" json:"orderId"`
	ChangeType string    `gorm:"column:changeType;type:enum('status','price','admin_remark');not null" json:"changeType"`
	OldValue   string    `gorm:"column:oldValue;size:500" json:"oldValue"`
	NewValue   string    `gorm:"column:newValue;size:500" json:"newValue"`
	Operator   string    `gorm:"size:100" json:"operator"`
	CreatedAt  time.Time `gorm:"column:createdAt" json:"createdAt"`
}

func (OutletOrderLog) TableName() string { return "outlet_order_logs" }

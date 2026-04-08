package models

import "time"

type Order struct {
	ID              int        `gorm:"primaryKey" json:"id"`
	OrderNo         string     `gorm:"column:orderNo;size:32;not null;uniqueIndex:orderNo" json:"orderNo"`
	UserID          int        `gorm:"column:userId;not null;index" json:"userId"`
	ServiceID       *int       `gorm:"column:serviceId" json:"serviceId"`
	ServiceTitle    string     `gorm:"column:serviceTitle;size:200;not null" json:"serviceTitle"`
	ServiceTitleEn  string     `gorm:"column:serviceTitleEn;size:200" json:"serviceTitleEn"`
	ServiceIcon     string     `gorm:"column:serviceIcon;size:100" json:"serviceIcon"`
	Price           float64    `gorm:"column:price;type:decimal(10,2)" json:"price"`
	Status          string     `gorm:"type:enum('pending','paid','processing','completed','cancelled');default:pending" json:"status"`
	Remark          string     `gorm:"type:text" json:"remark"`
	ContactName     string     `gorm:"column:contactName;size:50" json:"contactName"`
	ContactPhone    string     `gorm:"column:contactPhone;size:20" json:"contactPhone"`
	Address         string     `gorm:"size:500" json:"address"`
	AppointmentTime *time.Time `gorm:"column:appointmentTime" json:"appointmentTime"`
	AdminRemark     string     `gorm:"column:adminRemark;type:text" json:"adminRemark"`
	ProductSerial   string     `gorm:"column:productSerial;size:128" json:"productSerial"`
	GuideID         *int       `gorm:"column:guideId" json:"guideId"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	User            *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Guide           *DeviceGuide `gorm:"foreignKey:GuideID" json:"guide,omitempty"`
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

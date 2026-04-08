package models

import "time"

type User struct {
	ID          int        `gorm:"primaryKey" json:"id"`
	Username    string     `gorm:"size:50;not null;uniqueIndex:username" json:"username"`
	Email       *string    `gorm:"size:255;uniqueIndex:email" json:"email"`
	Password    string     `gorm:"size:255;not null" json:"-"`
	Nickname    string     `gorm:"size:100" json:"nickname"`
	Openid      *string    `gorm:"size:100;uniqueIndex:openid" json:"openid"`
	AlipayID    *string    `gorm:"column:alipayId;size:100;uniqueIndex:alipayId" json:"alipayId"`
	Avatar      string     `gorm:"size:500" json:"avatar"`
	Phone       string     `gorm:"size:20" json:"phone"`
	LastLoginIP *string    `gorm:"column:lastLoginIp;size:100" json:"lastLoginIp"`
	LastLoginAt *time.Time `gorm:"column:lastLoginAt" json:"lastLoginAt"`
	Role        string     `gorm:"type:enum('user','admin');default:user" json:"role"`
	Status      string     `gorm:"type:enum('active','disabled');default:active" json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	Messages    []Message  `gorm:"foreignKey:UserID" json:"messages,omitempty"`
	Addresses   []Address  `gorm:"foreignKey:UserID" json:"addresses,omitempty"`
}

func (User) TableName() string { return "users" }

type OutletUser struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"size:50;not null;uniqueIndex:outlet_users_username" json:"username"`
	Email     *string   `gorm:"size:255;uniqueIndex:outlet_users_email" json:"email"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	Nickname  string    `gorm:"size:100" json:"nickname"`
	Openid    *string   `gorm:"size:100;uniqueIndex:outlet_users_openid" json:"openid"`
	AlipayID  *string   `gorm:"column:alipayId;size:100;uniqueIndex:outlet_users_alipayId" json:"alipayId"`
	Avatar    string    `gorm:"size:500" json:"avatar"`
	Phone     string    `gorm:"size:20" json:"phone"`
	Role      string    `gorm:"type:enum('outlet','admin');default:outlet" json:"role"`
	Status    string    `gorm:"type:enum('active','disabled');default:active" json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Messages  []OutletMessage `gorm:"foreignKey:UserID" json:"messages,omitempty"`
	Addresses []OutletAddress `gorm:"foreignKey:UserID" json:"addresses,omitempty"`
}

func (OutletUser) TableName() string { return "outlet_users" }

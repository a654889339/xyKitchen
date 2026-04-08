package bootstrap

import (
	"log"

	"xykitchen/backend/internal/db"
	"xykitchen/backend/internal/models"
	"xykitchen/backend/internal/services"

	"gorm.io/gorm/clause"
)

const adminPassword = "XyKitchen@2026admin"

// Run 启动时：确保默认管理员存在（与 Node syncDatabase 一致）
func Run() error {
	var n int64
	if err := db.DB.Model(&models.User{}).Where("username = ?", "admin").Count(&n).Error; err != nil {
		return err
	}
	if n == 0 {
		hash, err := services.HashPassword(adminPassword)
		if err != nil {
			return err
		}
		email := "admin@xykitchen.local"
		u := models.User{
			Username: "admin",
			Email:    &email,
			Password: hash,
			Nickname: "管理员",
			Role:     "admin",
			Status:   "active",
		}
		if err := db.DB.Create(&u).Error; err != nil {
			return err
		}
		log.Println("[DB] Default admin account created.")
	}

	var i18nCount int64
	_ = db.DB.Model(&models.I18nText{}).Count(&i18nCount).Error
	if i18nCount == 0 {
		rows := []models.I18nText{
			{Key: "tabbar.home", Zh: "首页", En: "Home"},
			{Key: "tabbar.mine", Zh: "我的", En: "Mine"},
		}
		_ = db.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
		log.Println("[DB] Minimal i18n seed applied (full set may exist from previous Node deploy).")
	}

	return seedDefaultsIfEmpty()
}

func seedDefaultsIfEmpty() error {
	var pc int64
	if err := db.DB.Model(&models.ProductCategory{}).Count(&pc).Error; err != nil {
		return err
	}
	if pc == 0 {
		if err := db.DB.Create([]models.ProductCategory{
			{Name: "空调", SortOrder: 1, Status: "active"},
			{Name: "除湿与储能", SortOrder: 2, Status: "active"},
		}).Error; err != nil {
			return err
		}
		log.Println("[DB] Default product categories created.")
	}

	var hc int64
	if err := db.DB.Model(&models.HomeConfig{}).Count(&hc).Error; err != nil {
		return err
	}
	if hc == 0 {
		seed := []models.HomeConfig{
			{Section: "banner", Title: "Vino 品质服务", Desc: "专业·高效·可信赖", Color: "linear-gradient(135deg, #B91C1C, #7F1D1D)", SortOrder: 1, Status: "active"},
			{Section: "nav", Title: "全部服务", Icon: "apps-o", Path: "/services", Color: "#B91C1C", SortOrder: 1, Status: "active"},
		}
		if err := db.DB.Create(&seed).Error; err != nil {
			return err
		}
		log.Println("[DB] Default home configs (partial) created.")
	}

	var dg int64
	if err := db.DB.Model(&models.DeviceGuide{}).Count(&dg).Error; err != nil {
		return err
	}
	if dg == 0 {
		var pcs []models.ProductCategory
		if err := db.DB.Order("sortOrder ASC").Limit(2).Find(&pcs).Error; err != nil {
			return err
		}
		if len(pcs) >= 2 {
			c1, c2 := pcs[0].ID, pcs[1].ID
			guides := []models.DeviceGuide{
				{Name: "空调", Slug: strPtr("aircondition"), Subtitle: "家用/商用中央空调", Icon: "cluster-o", Emoji: "❄️", Gradient: "linear-gradient(135deg, #3B82F6, #1D4ED8)", Badge: "热门", SortOrder: 1, CategoryID: &c1, Status: "active"},
				{Name: "除湿机", Slug: strPtr("dehumidifier"), Subtitle: "家用/工业除湿设备", Icon: "filter-o", Emoji: "💧", Gradient: "linear-gradient(135deg, #06B6D4, #0891B2)", SortOrder: 2, CategoryID: &c2, Status: "active"},
			}
			if err := db.DB.Create(&guides).Error; err != nil {
				return err
			}
			log.Println("[DB] Default device guides (partial) created.")
		}
	}

	var bc int64
	if err := db.DB.Model(&models.BookingConfig{}).Count(&bc).Error; err != nil {
		return err
	}
	if bc == 0 {
		row := models.BookingConfig{
			NoticeTitle:      "預定須知",
			NoticeBody:       "請在小程序或網頁端完成預訂流程。",
			PerPersonDeposit: 50,
			TimeSlotsJSON:    `["17:00","19:00","21:00"]`,
			GuestOptionsJSON: `[2,3,4,6]`,
		}
		if err := db.DB.Create(&row).Error; err != nil {
			return err
		}
		log.Println("[DB] Default booking config created.")
	}

	return nil
}

func strPtr(s string) *string { return &s }

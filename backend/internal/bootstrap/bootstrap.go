package bootstrap

import (
	"log"

	"xykitchen/backend/internal/db"
	"xykitchen/backend/internal/models"
	"xykitchen/backend/internal/services"

	"gorm.io/gorm"
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

	// 若 i18n 为空，插入少量关键键（完整种子见原 Node 或数据库已有数据）
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

	var sc int64
	if err := db.DB.Model(&models.ServiceCategory{}).Count(&sc).Error; err != nil {
		return err
	}
	if sc == 0 {
		if err := db.DB.Create([]models.ServiceCategory{
			{Name: "维修", Key: strPtr("repair"), SortOrder: 1, Status: "active"},
			{Name: "清洁", Key: strPtr("clean"), SortOrder: 2, Status: "active"},
			{Name: "检测", Key: strPtr("inspect"), SortOrder: 3, Status: "active"},
			{Name: "数据", Key: strPtr("data"), SortOrder: 4, Status: "active"},
		}).Error; err != nil {
			return err
		}
		log.Println("[DB] Default service categories created.")
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

	var svc int64
	if err := db.DB.Model(&models.Service{}).Count(&svc).Error; err != nil {
		return err
	}
	if svc == 0 {
		var cats []models.ServiceCategory
		if err := db.DB.Order("sortOrder ASC").Find(&cats).Error; err != nil {
			return err
		}
		if len(cats) >= 4 {
			repair, clean, inspect, data := cats[0].ID, cats[1].ID, cats[2].ID, cats[3].ID
			services := []models.Service{
				{Title: "设备维修", Description: "专业工程师提供全方位维修服务", Icon: "setting-o", CategoryID: &repair, Price: 99, OriginPrice: f64Ptr(159), Bg: "#B91C1C", SortOrder: 1, Status: "active"},
				{Title: "深度清洁", Description: "全方位清洁保养", Icon: "brush-o", CategoryID: &clean, Price: 149, OriginPrice: f64Ptr(199), Bg: "#2563EB", SortOrder: 1, Status: "active"},
				{Title: "全面检测", Description: "系统全面评估", Icon: "scan", CategoryID: &inspect, Price: 49, OriginPrice: f64Ptr(79), Bg: "#059669", SortOrder: 1, Status: "active"},
				{Title: "数据恢复", Description: "专业数据找回", Icon: "replay", CategoryID: &data, Price: 199, OriginPrice: f64Ptr(299), Bg: "#7C3AED", SortOrder: 1, Status: "active"},
			}
			if err := db.DB.Create(&services).Error; err != nil {
				return err
			}
			log.Println("[DB] Default services created.")
		}
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

	return nil
}

func strPtr(s string) *string { return &s }
func f64Ptr(f float64) *float64 { return &f }

// CleanDuplicateIndexes 可选：与 Node 一致清理重复索引（简化版跳过）
func CleanDuplicateIndexes(_ *gorm.DB) {}

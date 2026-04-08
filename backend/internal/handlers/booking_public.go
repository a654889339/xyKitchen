package handlers

import (
	"encoding/json"
	"strconv"
	"time"

	"xykitchen/backend/internal/db"
	"xykitchen/backend/internal/models"
	"xykitchen/backend/internal/resp"

	"github.com/gin-gonic/gin"
)

func bookingCalendar(c *gin.Context) {
	y, _ := strconv.Atoi(c.Query("year"))
	m, _ := strconv.Atoi(c.Query("month"))
	if y < 2000 || m < 1 || m > 12 {
		resp.Err(c, 400, 1, "year/month 无效")
		return
	}
	loc := time.Local
	now := time.Now().In(loc)
	start := time.Date(y, time.Month(m), 1, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 1, -1)
	maxDay := end.Day()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	maxBook := today.AddDate(0, 0, 60)

	days := make([]int, 0, 32)
	for d := 1; d <= maxDay; d++ {
		dt := time.Date(y, time.Month(m), d, 0, 0, 0, 0, loc)
		if dt.Before(today) {
			continue
		}
		if dt.After(maxBook) {
			continue
		}
		days = append(days, d)
	}
	resp.OK(c, gin.H{"year": y, "month": m, "days": days})
}

func bookingMeta(c *gin.Context) {
	var cfg models.BookingConfig
	_ = db.DB.Order("id ASC").First(&cfg).Error
	slots := []string{"17:00", "19:00", "21:00"}
	if cfg.TimeSlotsJSON != "" {
		_ = json.Unmarshal([]byte(cfg.TimeSlotsJSON), &slots)
	}
	guests := []int{2, 3, 4, 6}
	if cfg.GuestOptionsJSON != "" {
		_ = json.Unmarshal([]byte(cfg.GuestOptionsJSON), &guests)
	}
	deposit := cfg.PerPersonDeposit
	if deposit <= 0 {
		deposit = 50
	}
	resp.OK(c, gin.H{
		"timeSlots":        slots,
		"guestOptions":     guests,
		"perPersonDeposit": deposit,
	})
}

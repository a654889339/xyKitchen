package handlers

import (
	"encoding/json"
	"io"
	"path"
	"strconv"
	"strings"
	"time"

	"xykitchen/backend/internal/config"
	"xykitchen/backend/internal/db"
	"xykitchen/backend/internal/models"
	"xykitchen/backend/internal/resp"
	"xykitchen/backend/internal/services"

	"github.com/gin-gonic/gin"
)

func bookingConfigGet(c *gin.Context) {
	var cfg models.BookingConfig
	err := db.DB.Order("id ASC").First(&cfg).Error
	if err != nil {
		resp.OK(c, gin.H{
			"homepageBgUrl":    "",
			"noticeTitle":      "",
			"noticeBody":       "",
			"perPersonDeposit": 50,
			"timeSlots":        []string{"17:00", "19:00", "21:00"},
			"guestOptions":     []int{2, 3, 4, 6},
		})
		return
	}
	slots := []string{"17:00", "19:00", "21:00"}
	if cfg.TimeSlotsJSON != "" {
		_ = json.Unmarshal([]byte(cfg.TimeSlotsJSON), &slots)
	}
	guests := []int{2, 3, 4, 6}
	if cfg.GuestOptionsJSON != "" {
		_ = json.Unmarshal([]byte(cfg.GuestOptionsJSON), &guests)
	}
	resp.OK(c, gin.H{
		"id":               cfg.ID,
		"homepageBgUrl":    fixHomeProxyURL(cfg.HomepageBgURL),
		"noticeTitle":      cfg.NoticeTitle,
		"noticeBody":       cfg.NoticeBody,
		"perPersonDeposit": cfg.PerPersonDeposit,
		"timeSlots":        slots,
		"guestOptions":     guests,
		"updatedAt":        cfg.UpdatedAt,
	})
}

func bookingConfigPut(c *gin.Context) {
	var body struct {
		NoticeTitle      *string   `json:"noticeTitle"`
		NoticeBody       *string   `json:"noticeBody"`
		PerPersonDeposit *float64  `json:"perPersonDeposit"`
		TimeSlots        []string  `json:"timeSlots"`
		GuestOptions     []int     `json:"guestOptions"`
		HomepageBgURL    *string   `json:"homepageBgUrl"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Err(c, 400, 1, err.Error())
		return
	}
	var cfg models.BookingConfig
	err := db.DB.Order("id ASC").First(&cfg).Error
	isNew := err != nil
	if isNew {
		cfg = models.BookingConfig{
			PerPersonDeposit: 50,
			TimeSlotsJSON:    `["17:00","19:00","21:00"]`,
			GuestOptionsJSON: `[2,3,4,6]`,
		}
	}
	if body.NoticeTitle != nil {
		cfg.NoticeTitle = strings.TrimSpace(*body.NoticeTitle)
	}
	if body.NoticeBody != nil {
		cfg.NoticeBody = *body.NoticeBody
	}
	if body.PerPersonDeposit != nil && *body.PerPersonDeposit >= 0 {
		cfg.PerPersonDeposit = *body.PerPersonDeposit
	}
	if body.HomepageBgURL != nil {
		cfg.HomepageBgURL = strings.TrimSpace(*body.HomepageBgURL)
	}
	if len(body.TimeSlots) > 0 {
		b, _ := json.Marshal(body.TimeSlots)
		cfg.TimeSlotsJSON = string(b)
	}
	if len(body.GuestOptions) > 0 {
		b, _ := json.Marshal(body.GuestOptions)
		cfg.GuestOptionsJSON = string(b)
	}
	if isNew {
		if err := db.DB.Create(&cfg).Error; err != nil {
			resp.Err(c, 500, 1, err.Error())
			return
		}
	} else {
		if err := db.DB.Save(&cfg).Error; err != nil {
			resp.Err(c, 500, 1, err.Error())
			return
		}
	}
	bookingConfigGet(c)
}

func bookingConfigUploadBg(c *gin.Context, cfg *config.Config) {
	_ = cfg
	fh, err := c.FormFile("file")
	if err != nil {
		resp.Err(c, 400, 1, "请选择图片文件")
		return
	}
	f, err := fh.Open()
	if err != nil {
		resp.Err(c, 500, 1, err.Error())
		return
	}
	defer f.Close()
	buf, err := io.ReadAll(f)
	if err != nil {
		resp.Err(c, 500, 1, err.Error())
		return
	}
	ext := path.Ext(fh.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := "booking-home-" + strconv.FormatInt(time.Now().UnixMilli(), 10) + ext
	ct := fh.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/jpeg"
	}
	urlu, _, err := services.UploadWithThumb(c.Request.Context(), buf, filename, ct, 0)
	if err != nil {
		resp.Err(c, 500, 1, err.Error())
		return
	}
	var row models.BookingConfig
	if err := db.DB.Order("id ASC").First(&row).Error; err != nil {
		row = models.BookingConfig{
			HomepageBgURL:    urlu,
			NoticeTitle:      "預定須知",
			PerPersonDeposit: 50,
			TimeSlotsJSON:    `["17:00","19:00","21:00"]`,
			GuestOptionsJSON: `[2,3,4,6]`,
		}
		_ = db.DB.Create(&row).Error
	} else {
		row.HomepageBgURL = urlu
		_ = db.DB.Save(&row).Error
	}
	resp.OK(c, gin.H{"url": urlu, "homepageBgUrl": fixHomeProxyURL(urlu)})
}

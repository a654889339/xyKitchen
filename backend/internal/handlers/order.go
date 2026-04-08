package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"xykitchen/backend/internal/db"
	"xykitchen/backend/internal/models"
	"xykitchen/backend/internal/resp"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var phone11 = regexp.MustCompile(`^1[3-9]\d{9}$`)

var orderStatusMap = map[string]struct {
	Text   string `json:"text"`
	TextEn string `json:"textEn"`
	Type   string `json:"type"`
}{
	"pending":    {Text: "待支付", TextEn: "Unpaid", Type: "warning"},
	"paid":       {Text: "已支付", TextEn: "Paid", Type: "primary"},
	"processing": {Text: "进行中", TextEn: "In Progress", Type: "primary"},
	"completed":  {Text: "已完成", TextEn: "Completed", Type: "success"},
	"cancelled":  {Text: "已取消", TextEn: "Cancelled", Type: "default"},
}

func genOrderNo() string {
	now := time.Now()
	r := rand.Intn(10000)
	return fmt.Sprintf("XK%d%02d%02d%02d%02d%02d%04d",
		now.Year(), int(now.Month()), now.Day(), now.Hour(), now.Minute(), now.Second(), r)
}

func parseTimeSlot(slot string) (hour, min int, ok bool) {
	slot = strings.TrimSpace(slot)
	parts := strings.Split(slot, ":")
	if len(parts) != 2 {
		return 0, 0, false
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, 0, false
	}
	return h, m, true
}

func orderCreate(c *gin.Context) {
	u, ok := ctxUser(c)
	if !ok {
		return
	}
	var body struct {
		BookingDate  string  `json:"bookingDate"`
		TimeSlot     string  `json:"timeSlot"`
		GuestCount   int     `json:"guestCount"`
		ContactPhone string  `json:"contactPhone"`
		Price        float64 `json:"price"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Err(c, 400, 400, "参数错误")
		return
	}
	phone := strings.TrimSpace(body.ContactPhone)
	if !phone11.MatchString(phone) {
		resp.Err(c, 400, 400, "请输入11位大陆手机号")
		return
	}
	if body.GuestCount < 1 || body.GuestCount > 50 {
		resp.Err(c, 400, 400, "用餐人数无效")
		return
	}
	var bcfg models.BookingConfig
	_ = db.DB.Order("id ASC").First(&bcfg).Error
	per := bcfg.PerPersonDeposit
	if per <= 0 {
		per = 50
	}
	expect := per * float64(body.GuestCount)
	if math.Abs(expect-body.Price) > 0.02 {
		resp.Err(c, 400, 400, "金额与订金规则不符，请刷新后重试")
		return
	}
	ds := strings.TrimSpace(body.BookingDate)
	d, err := time.ParseInLocation("2006-01-02", ds, time.Local)
	if err != nil {
		resp.Err(c, 400, 400, "预订日期无效")
		return
	}
	h, mi, tok := parseTimeSlot(body.TimeSlot)
	if !tok {
		resp.Err(c, 400, 400, "用餐时段无效")
		return
	}
	bk := time.Date(d.Year(), d.Month(), d.Day(), h, mi, 0, 0, time.Local)
	if !bk.After(time.Now().Add(-1 * time.Minute)) {
		resp.Err(c, 400, 400, "预订时间需晚于当前时间")
		return
	}
	bkp := bk
	o := models.Order{
		OrderNo:      genOrderNo(),
		UserID:       u.ID,
		BookingAt:    &bkp,
		GuestCount:   body.GuestCount,
		ContactPhone: phone,
		Price:        expect,
		Status:       "pending",
		ServiceTitle: "门店预订",
	}
	if err := db.DB.Create(&o).Error; err != nil {
		resp.Err(c, 500, 500, "创建订单失败")
		return
	}
	resp.OK(c, o)
}

func orderMyOrders(c *gin.Context) {
	u, ok := ctxUser(c)
	if !ok {
		return
	}
	status := c.Query("status")
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "pageSize", 10)
	if pageSize > 100 {
		pageSize = 100
	}
	qb := db.DB.Model(&models.Order{}).Where("userId = ?", u.ID)
	if status != "" && status != "all" {
		qb = qb.Where("status = ?", status)
	}
	var total int64
	qb.Count(&total)
	var rows []models.Order
	qb.Order("createdAt DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&rows)
	list := make([]gin.H, 0, len(rows))
	for _, o := range rows {
		s := orderStatusMap[o.Status]
		if s.Text == "" {
			s = orderStatusMap["pending"]
		}
		raw, _ := json.Marshal(o)
		var h gin.H
		_ = json.Unmarshal(raw, &h)
		h["statusText"] = s.Text
		h["statusTextEn"] = s.TextEn
		h["statusType"] = s.Type
		list = append(list, h)
	}
	resp.OK(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

func orderDetail(c *gin.Context) {
	u, ok := ctxUser(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		resp.Err(c, 400, 400, "无效订单")
		return
	}
	var o models.Order
	if err := db.DB.First(&o, id).Error; err != nil {
		resp.Err(c, 404, 404, "订单不存在")
		return
	}
	if o.UserID != u.ID && u.Role != "admin" {
		resp.Err(c, 403, 403, "无权查看")
		return
	}
	s := orderStatusMap[o.Status]
	raw, _ := json.Marshal(o)
	var h gin.H
	_ = json.Unmarshal(raw, &h)
	h["statusText"] = s.Text
	h["statusTextEn"] = s.TextEn
	h["statusType"] = s.Type
	resp.OK(c, h)
}

func orderCancel(c *gin.Context) {
	u, ok := ctxUser(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		resp.Err(c, 400, 400, "无效订单")
		return
	}
	var o models.Order
	if err := db.DB.First(&o, id).Error; err != nil {
		resp.Err(c, 404, 404, "订单不存在")
		return
	}
	if o.UserID != u.ID && u.Role != "admin" {
		resp.Err(c, 403, 403, "无权操作")
		return
	}
	if o.Status == "completed" || o.Status == "cancelled" {
		resp.Err(c, 400, 400, "当前状态无法取消")
		return
	}
	o.Status = "cancelled"
	db.DB.Save(&o)
	resp.OKMsg(c, "订单已取消")
}

func orderAdminList(c *gin.Context) {
	status := c.Query("status")
	orderNo := strings.TrimSpace(c.Query("orderNo"))
	userID := strings.TrimSpace(c.Query("userId"))
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "pageSize", 50)
	if pageSize > 200 {
		pageSize = 200
	}
	qb := db.DB.Model(&models.Order{})
	if status != "" && status != "all" {
		qb = qb.Where("status = ?", status)
	}
	if orderNo != "" {
		qb = qb.Where("orderNo LIKE ?", "%"+escapeLike(orderNo)+"%")
	}
	if userID != "" {
		if uid, err := strconv.Atoi(userID); err == nil && uid > 0 {
			qb = qb.Where("userId = ?", uid)
		}
	}
	var total int64
	sq := qb.Session(&gorm.Session{})
	sq.Count(&total)
	var rows []models.Order
	fq := qb.Session(&gorm.Session{})
	fq.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "username", "email", "nickname", "phone")
	}).Order("createdAt DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&rows)
	list := make([]gin.H, 0, len(rows))
	for _, o := range rows {
		s := orderStatusMap[o.Status]
		raw, _ := json.Marshal(o)
		var h gin.H
		_ = json.Unmarshal(raw, &h)
		h["statusText"] = s.Text
		h["statusTextEn"] = s.TextEn
		h["statusType"] = s.Type
		list = append(list, h)
	}
	resp.OK(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

func orderAdminStats(c *gin.Context) {
	var total, pending, processing, completed, cancelled int64
	db.DB.Model(&models.Order{}).Count(&total)
	db.DB.Model(&models.Order{}).Where("status = ?", "pending").Count(&pending)
	db.DB.Model(&models.Order{}).Where("status = ?", "processing").Count(&processing)
	db.DB.Model(&models.Order{}).Where("status = ?", "completed").Count(&completed)
	db.DB.Model(&models.Order{}).Where("status = ?", "cancelled").Count(&cancelled)
	resp.OK(c, gin.H{"total": total, "pending": pending, "processing": processing, "completed": completed, "cancelled": cancelled})
}

func orderAdminUpdateStatus(c *gin.Context) {
	u, ok := ctxUser(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		resp.Err(c, 400, 400, "无效订单")
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Err(c, 400, 400, "无效状态")
		return
	}
	if _, ok := orderStatusMap[body.Status]; !ok {
		resp.Err(c, 400, 400, "无效状态")
		return
	}
	var o models.Order
	if err := db.DB.First(&o, id).Error; err != nil {
		resp.Err(c, 404, 404, "订单不存在")
		return
	}
	old := o.Status
	if old != body.Status {
		oldS := orderStatusMap[old].Text
		newS := orderStatusMap[body.Status].Text
		db.DB.Create(&models.OrderLog{
			OrderID:    o.ID,
			ChangeType: "status",
			OldValue:   firstNonEmptyStr(oldS, old),
			NewValue:   firstNonEmptyStr(newS, body.Status),
			Operator:   u.Username,
		})
	}
	o.Status = body.Status
	db.DB.Save(&o)
	s := orderStatusMap[o.Status]
	raw, _ := json.Marshal(o)
	var h gin.H
	_ = json.Unmarshal(raw, &h)
	h["statusText"] = s.Text
	h["statusTextEn"] = s.TextEn
	h["statusType"] = s.Type
	resp.OK(c, h)
}

func orderAdminUpdatePrice(c *gin.Context) {
	u, ok := ctxUser(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		resp.Err(c, 400, 400, "无效订单")
		return
	}
	var body struct {
		Price float64 `json:"price"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Err(c, 400, 400, "无效金额")
		return
	}
	if body.Price < 0 || math.IsNaN(body.Price) {
		resp.Err(c, 400, 400, "无效金额")
		return
	}
	var o models.Order
	if err := db.DB.First(&o, id).Error; err != nil {
		resp.Err(c, 404, 404, "订单不存在")
		return
	}
	oldP := o.Price
	if oldP != body.Price {
		db.DB.Create(&models.OrderLog{
			OrderID:    o.ID,
			ChangeType: "price",
			OldValue:   fmt.Sprintf("¥%.2f", oldP),
			NewValue:   fmt.Sprintf("¥%.2f", body.Price),
			Operator:   u.Username,
		})
	}
	o.Price = body.Price
	db.DB.Save(&o)
	resp.OK(c, o)
}

func orderAdminAddRemark(c *gin.Context) {
	u, ok := ctxUser(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		resp.Err(c, 400, 400, "无效订单")
		return
	}
	var body struct {
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Remark) == "" {
		resp.Err(c, 400, 400, "备注不能为空")
		return
	}
	var o models.Order
	if err := db.DB.First(&o, id).Error; err != nil {
		resp.Err(c, 404, 404, "订单不存在")
		return
	}
	db.DB.Create(&models.OrderLog{
		OrderID:    o.ID,
		ChangeType: "admin_remark",
		OldValue:   "",
		NewValue:   strings.TrimSpace(body.Remark),
		Operator:   u.Username,
	})
	o.AdminRemark = strings.TrimSpace(body.Remark)
	db.DB.Save(&o)
	resp.OKMsg(c, "备注已添加")
}

func orderAdminLogs(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		resp.Err(c, 400, 400, "无效订单")
		return
	}
	var logs []models.OrderLog
	db.DB.Where("orderId = ?", id).Order("createdAt DESC").Find(&logs)
	var o models.Order
	db.DB.Select("id", "orderNo", "adminRemark").First(&o, id)
	resp.OK(c, gin.H{"logs": logs, "adminRemark": o.AdminRemark})
}

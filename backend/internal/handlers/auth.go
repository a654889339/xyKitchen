package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"xykitchen/backend/internal/config"
	"xykitchen/backend/internal/db"
	"xykitchen/backend/internal/models"
	"xykitchen/backend/internal/resp"
	"xykitchen/backend/internal/services"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	zxing "github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"gorm.io/gorm"
)

var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func authSendCode(c *gin.Context, cfg *config.Config) {
	var body struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Err(c, 400, 400, "邮箱不能为空")
		return
	}
	email := strings.TrimSpace(body.Email)
	if email == "" || !emailRegex.MatchString(email) {
		resp.Err(c, 400, 400, "邮箱格式不正确")
		return
	}
	var n int64
	if err := db.DB.Model(&models.User{}).Where("email = ?", email).Count(&n).Error; err != nil {
		resp.Err(c, 500, 500, "服务器错误")
		return
	}
	if n > 0 {
		resp.Err(c, 400, 400, "该邮箱已被注册")
		return
	}
	if err := services.SendEmailCode(cfg, email); err != nil {
		resp.Err(c, 400, 400, err.Error())
		return
	}
	resp.OKMsg(c, "验证码已发送")
}

func authSendSmsCode(c *gin.Context, cfg *config.Config) {
	var body struct {
		Phone string `json:"phone"`
		Scene string `json:"scene"`
	}
	_ = c.ShouldBindJSON(&body)
	phone := strings.TrimSpace(body.Phone)
	if phone == "" {
		resp.Err(c, 400, 400, "手机号不能为空")
		return
	}
	key := services.NormalizePhone(phone)
	if len(key) != 11 || key[0] != '1' {
		resp.Err(c, 400, 400, "请输入正确的11位大陆手机号")
		return
	}
	if body.Scene == "register" {
		var n int64
		_ = db.DB.Model(&models.User{}).Where("phone = ?", key).Count(&n).Error
		if n > 0 {
			resp.Err(c, 400, 400, "该手机号已注册")
			return
		}
	}
	if err := services.SendSMSCode(cfg, phone); err != nil {
		resp.Err(c, 400, 400, err.Error())
		return
	}
	resp.OKMsg(c, "验证码已发送")
}

func authRegister(c *gin.Context, cfg *config.Config) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    *string
		Code     string `json:"code"`
		Nickname string `json:"nickname"`
		Phone    string `json:"phone"`
		SmsCode  string `json:"smsCode"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Err(c, 400, 400, "参数错误")
		return
	}
	usePhone := strings.TrimSpace(body.Phone) != ""
	if usePhone {
		code := body.SmsCode
		if code == "" {
			code = body.Code
		}
		if code == "" {
			resp.Err(c, 400, 400, "验证码不能为空")
			return
		}
		normalized := services.NormalizePhone(body.Phone)
		if len(normalized) != 11 || normalized[0] != '1' {
			resp.Err(c, 400, 400, "手机号格式不正确")
			return
		}
		ok, msg := services.SMSVerify(body.Phone, code)
		if !ok {
			resp.Err(c, 400, 400, msg)
			return
		}
		var n int64
		_ = db.DB.Model(&models.User{}).Where("phone = ?", normalized).Count(&n).Error
		if n > 0 {
			resp.Err(c, 400, 400, "该手机号已注册")
			return
		}
		base := strings.TrimSpace(body.Username)
		if base == "" {
			base = "u" + normalized[len(normalized)-8:]
		}
		if strings.EqualFold(base, "admin") {
			resp.Err(c, 400, 400, "该用户名为系统保留，不可注册")
			return
		}
		finalName := base
		for i := 0; ; i++ {
			var cnt int64
			db.DB.Model(&models.User{}).Where("username = ?", finalName).Count(&cnt)
			if cnt == 0 {
				break
			}
			finalName = fmt.Sprintf("%s%d", base, i+1)
		}
		if len(body.Password) < 6 {
			resp.Err(c, 400, 400, "密码长度不能少于6位")
			return
		}
		hash, err := services.HashPassword(body.Password)
		if err != nil {
			resp.Err(c, 500, 500, "注册失败")
			return
		}
		nick := strings.TrimSpace(body.Nickname)
		if nick == "" {
			nick = normalized[:3] + "****" + normalized[len(normalized)-4:]
		}
		u := models.User{Username: finalName, Password: hash, Phone: normalized, Nickname: nick, Role: "user", Status: "active"}
		if err := db.DB.Create(&u).Error; err != nil {
			resp.Err(c, 500, 500, "注册失败，请稍后重试")
			return
		}
		tok, _ := services.SignJWT(cfg, u.ID, u.Username, u.Role, "")
		resp.OK(c, gin.H{"token": tok, "user": u})
		return
	}
	if strings.TrimSpace(body.Username) == "" || body.Password == "" {
		resp.Err(c, 400, 400, "用户名和密码不能为空")
		return
	}
	if body.Email == nil || strings.TrimSpace(*body.Email) == "" {
		resp.Err(c, 400, 400, "邮箱不能为空")
		return
	}
	if body.Code == "" {
		resp.Err(c, 400, 400, "验证码不能为空")
		return
	}
	uName := strings.TrimSpace(body.Username)
	if len(uName) < 2 || len(uName) > 50 {
		resp.Err(c, 400, 400, "用户名长度需在2-50个字符之间")
		return
	}
	if len(body.Password) < 6 {
		resp.Err(c, 400, 400, "密码长度不能少于6位")
		return
	}
	em := strings.TrimSpace(*body.Email)
	ok, msg := services.EmailVerify(em, body.Code)
	if !ok {
		resp.Err(c, 400, 400, msg)
		return
	}
	if strings.EqualFold(uName, "admin") {
		resp.Err(c, 400, 400, "该用户名为系统保留，不可注册")
		return
	}
	var cnt int64
	db.DB.Model(&models.User{}).Where("username = ?", uName).Count(&cnt)
	if cnt > 0 {
		resp.Err(c, 400, 400, "用户名已存在")
		return
	}
	db.DB.Model(&models.User{}).Where("email = ?", em).Count(&cnt)
	if cnt > 0 {
		resp.Err(c, 400, 400, "该邮箱已被注册")
		return
	}
	hash, err := services.HashPassword(body.Password)
	if err != nil {
		resp.Err(c, 500, 500, "注册失败")
		return
	}
	nick := strings.TrimSpace(body.Nickname)
	if nick == "" {
		nick = uName
	}
	emailPtr := em
	u := models.User{Username: uName, Password: hash, Email: &emailPtr, Nickname: nick, Role: "user", Status: "active"}
	if err := db.DB.Create(&u).Error; err != nil {
		resp.Err(c, 500, 500, "注册失败，请稍后重试")
		return
	}
	tok, _ := services.SignJWT(cfg, u.ID, u.Username, u.Role, "")
	resp.OK(c, gin.H{"token": tok, "user": u})
}

func authLogin(c *gin.Context, cfg *config.Config) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Phone    string `json:"phone"`
		Code     string `json:"code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Err(c, 400, 400, "参数错误")
		return
	}
	if strings.TrimSpace(body.Phone) != "" {
		if body.Code == "" {
			resp.Err(c, 400, 400, "验证码不能为空")
			return
		}
		normalized := services.NormalizePhone(body.Phone)
		if len(normalized) != 11 || normalized[0] != '1' {
			resp.Err(c, 400, 400, "手机号格式不正确")
			return
		}
		ok, msg := services.SMSVerify(body.Phone, body.Code)
		if !ok {
			resp.Err(c, 400, 400, msg)
			return
		}
		var u models.User
		if err := db.DB.Where("phone = ?", normalized).First(&u).Error; err != nil {
			resp.Err(c, 400, 400, "该手机号未注册，请先注册")
			return
		}
		if u.Status != "active" {
			resp.Err(c, 403, 403, "账号已被禁用")
			return
		}
		touchUserLogin(c, &u)
		tok, _ := services.SignJWT(cfg, u.ID, u.Username, u.Role, "")
		resp.OK(c, gin.H{"token": tok, "user": u})
		return
	}
	if strings.TrimSpace(body.Username) == "" || body.Password == "" {
		resp.Err(c, 400, 400, "用户名和密码不能为空")
		return
	}
	var u models.User
	if err := db.DB.Where("username = ?", strings.TrimSpace(body.Username)).First(&u).Error; err != nil {
		resp.Err(c, 401, 401, "用户名或密码错误")
		return
	}
	if !services.CheckPassword(u.Password, body.Password) {
		resp.Err(c, 401, 401, "用户名或密码错误")
		return
	}
	if u.Status != "active" {
		resp.Err(c, 403, 403, "账号已被禁用")
		return
	}
	touchUserLogin(c, &u)
	tok, _ := services.SignJWT(cfg, u.ID, u.Username, u.Role, "")
	resp.OK(c, gin.H{"token": tok, "user": u})
}

func touchUserLogin(c *gin.Context, u *models.User) {
	ip := c.ClientIP()
	if xf := c.GetHeader("X-Forwarded-For"); xf != "" {
		ip = strings.TrimSpace(strings.Split(xf, ",")[0])
	}
	u.LastLoginIP = &ip
	now := time.Now()
	u.LastLoginAt = &now
	_ = db.DB.Model(u).Updates(map[string]interface{}{"lastLoginIp": ip, "lastLoginAt": now}).Error
}

func authBindPhone(c *gin.Context, cfg *config.Config) {
	u, ok := ctxUser(c)
	if !ok {
		return
	}
	var body struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Phone == "" || body.Code == "" {
		resp.Err(c, 400, 400, "手机号和验证码不能为空")
		return
	}
	normalized := services.NormalizePhone(body.Phone)
	if len(normalized) != 11 || normalized[0] != '1' {
		resp.Err(c, 400, 400, "手机号格式不正确")
		return
	}
	ok2, msg := services.SMSVerify(body.Phone, body.Code)
	if !ok2 {
		resp.Err(c, 400, 400, msg)
		return
	}
	var other models.User
	if err := db.DB.Where("phone = ?", normalized).First(&other).Error; err == nil && other.ID != u.ID {
		resp.Err(c, 400, 400, "该手机号已被其他账号绑定")
		return
	}
	var user models.User
	if err := db.DB.First(&user, u.ID).Error; err != nil {
		resp.Err(c, 404, 404, "用户不存在")
		return
	}
	user.Phone = normalized
	if err := db.DB.Model(&user).Update("phone", normalized).Error; err != nil {
		resp.Err(c, 500, 500, err.Error())
		return
	}
	resp.JSON(c, 0, gin.H{"data": user, "message": "绑定成功"})
}

func authWxLogin(c *gin.Context, cfg *config.Config) {
	var body struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Code == "" {
		resp.Err(c, 400, 400, "code不能为空")
		return
	}
	openid, err := services.WxCode2Session(cfg, body.Code)
	if err != nil {
		resp.Err(c, 400, 400, err.Error())
		return
	}
	var user models.User
	err = db.DB.Where("openid = ?", openid).First(&user).Error
	isNew := false
	if err == gorm.ErrRecordNotFound {
		b := make([]byte, 4)
		_, _ = rand.Read(b)
		short := hex.EncodeToString(b)
		randPwd := make([]byte, 16)
		_, _ = rand.Read(randPwd)
		hash, _ := services.HashPassword(hex.EncodeToString(randPwd))
		user = models.User{
			Username:     "wx_" + short,
			Password:     hash,
			Nickname:     "微信用户",
			Openid:       &openid,
			Role:         "user",
			Status:       "active",
		}
		if err := db.DB.Create(&user).Error; err != nil {
			resp.Err(c, 500, 500, "微信登录失败")
			return
		}
		isNew = true
	} else if err != nil {
		resp.Err(c, 500, 500, "微信登录失败")
		return
	}
	touchUserLogin(c, &user)
	tok, _ := services.SignJWT(cfg, user.ID, user.Username, user.Role, "")
	resp.OK(c, gin.H{"token": tok, "user": user, "isNew": isNew})
}

func authAlipayLogin(c *gin.Context, cfg *config.Config) {
	var body struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Code == "" {
		resp.Err(c, 400, 400, "code不能为空")
		return
	}
	uid, err := services.AlipayOAuthToken(cfg, body.Code)
	if err != nil {
		resp.Err(c, 500, 500, "支付宝登录失败")
		return
	}
	var user models.User
	err = db.DB.Where("alipayId = ?", uid).First(&user).Error
	isNew := false
	if err == gorm.ErrRecordNotFound {
		b := make([]byte, 4)
		_, _ = rand.Read(b)
		short := hex.EncodeToString(b)
		randPwd := make([]byte, 16)
		_, _ = rand.Read(randPwd)
		hash, _ := services.HashPassword(hex.EncodeToString(randPwd))
		alipayID := uid
		user = models.User{
			Username: "ali_" + short,
			Password: hash,
			Nickname: "支付宝用户",
			AlipayID: &alipayID,
			Role:     "user",
			Status:   "active",
		}
		if err := db.DB.Create(&user).Error; err != nil {
			resp.Err(c, 500, 500, "支付宝登录失败")
			return
		}
		isNew = true
	} else if err != nil {
		resp.Err(c, 500, 500, "支付宝登录失败")
		return
	}
	touchUserLogin(c, &user)
	tok, _ := services.SignJWT(cfg, user.ID, user.Username, user.Role, "")
	resp.OK(c, gin.H{"token": tok, "user": user, "isNew": isNew})
}

func authGetProfile(c *gin.Context) {
	u, ok := ctxUser(c)
	if !ok {
		return
	}
	var user models.User
	if err := db.DB.First(&user, u.ID).Error; err != nil {
		resp.Err(c, 404, 404, "用户不存在")
		return
	}
	if user.Status != "active" {
		resp.Err(c, 403, 403, "账号已被禁用")
		return
	}
	resp.OK(c, user)
}

func authUpdateProfile(c *gin.Context) {
	u, ok := ctxUser(c)
	if !ok {
		return
	}
	var body struct {
		Nickname *string `json:"nickname"`
		Avatar   *string `json:"avatar"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Err(c, 400, 400, "参数错误")
		return
	}
	var user models.User
	if err := db.DB.First(&user, u.ID).Error; err != nil {
		resp.Err(c, 404, 404, "用户不存在")
		return
	}
	if body.Nickname != nil {
		s := strings.TrimSpace(*body.Nickname)
		if s != "" {
			user.Nickname = s
		}
	}
	if body.Avatar != nil {
		user.Avatar = strings.TrimSpace(*body.Avatar)
	}
	if err := db.DB.Save(&user).Error; err != nil {
		resp.Err(c, 500, 500, "更新失败")
		return
	}
	resp.OK(c, user)
}

func authUploadAvatar(c *gin.Context, cfg *config.Config) {
	u, ok := ctxUser(c)
	if !ok {
		return
	}
	fh, err := c.FormFile("avatar")
	if err != nil {
		resp.Err(c, 400, 400, "请选择图片")
		return
	}
	f, err := fh.Open()
	if err != nil {
		resp.Err(c, 400, 400, "读取失败")
		return
	}
	defer f.Close()
	buf, err := io.ReadAll(f)
	if err != nil {
		resp.Err(c, 500, 500, "读取失败")
		return
	}
	ext := path.Ext(fh.Filename)
	if ext == "" {
		ext = ".png"
	}
	rnd := make([]byte, 8)
	_, _ = rand.Read(rnd)
	filename := "avatar_" + hex.EncodeToString(rnd) + ext
	ct := fh.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/png"
	}
	url, err := services.UploadCOS(c.Request.Context(), buf, filename, ct)
	if err != nil {
		resp.Err(c, 500, 500, "上传失败")
		return
	}
	_ = db.DB.Model(&models.User{}).Where("id = ?", u.ID).Update("avatar", url).Error
	resp.OK(c, gin.H{"url": url})
}

func authAdminGetUsers(c *gin.Context) {
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "pageSize", 50)
	if pageSize > 200 {
		pageSize = 200
	}
	q := strings.TrimSpace(c.Query("q"))
	searchType := c.Query("searchType")

	qb := db.DB.Model(&models.User{})
	if q != "" && searchType != "" {
		switch searchType {
		case "id":
			id, _ := strconv.Atoi(q)
			if id > 0 {
				qb = qb.Where("id = ?", id)
			} else {
				qb = qb.Where("id = ?", -1)
			}
		case "username":
			qb = qb.Where("username LIKE ?", "%"+escapeLike(q)+"%")
		case "phone":
			qb = qb.Where("phone LIKE ?", "%"+escapeLike(q)+"%")
		}
	}

	var total int64
	sq := qb.Session(&gorm.Session{})
	if err := sq.Count(&total).Error; err != nil {
		resp.Err(c, 500, 500, "获取用户列表失败")
		return
	}
	var users []models.User
	offset := (page - 1) * pageSize
	fq := qb.Session(&gorm.Session{})
	if err := fq.Preload("Addresses").Order("createdAt DESC").Limit(pageSize).Offset(offset).Find(&users).Error; err != nil {
		resp.Err(c, 500, 500, "获取用户列表失败")
		return
	}
	ids := make([]int, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	orderCounts := map[int]int{}
	orderIdsByUser := map[int][]int{}
	boundByUser := map[int][]string{}
	if len(ids) > 0 {
		type cntRow struct {
			UserID int `gorm:"column:userId"`
			Cnt    int `gorm:"column:cnt"`
		}
		var crows []cntRow
		_ = db.DB.Model(&models.Order{}).Select("userId, COUNT(id) as cnt").Where("userId IN ?", ids).Group("userId").Scan(&crows).Error
		for _, r := range crows {
			orderCounts[r.UserID] = r.Cnt
		}
		var ordRows []models.Order
		db.DB.Select("id", "userId").Where("userId IN ?", ids).Order("id ASC").Find(&ordRows)
		for _, o := range ordRows {
			orderIdsByUser[o.UserID] = append(orderIdsByUser[o.UserID], o.ID)
		}
		var ups []models.UserProduct
		db.DB.Where("userId IN ?", ids).Find(&ups)
		for _, p := range ups {
			boundByUser[p.UserID] = append(boundByUser[p.UserID], p.ProductKey)
		}
	}
	list := make([]gin.H, 0, len(users))
	for _, u := range users {
		raw, _ := json.Marshal(u)
		var plain gin.H
		_ = json.Unmarshal(raw, &plain)
		delete(plain, "password")
		plain["orderCount"] = orderCounts[u.ID]
		plain["orderIds"] = orderIdsByUser[u.ID]
		if plain["orderIds"] == nil {
			plain["orderIds"] = []int{}
		}
		plain["boundProductKeys"] = boundByUser[u.ID]
		if plain["boundProductKeys"] == nil {
			plain["boundProductKeys"] = []string{}
		}
		list = append(list, plain)
	}
	var totalUsers, adminCount, totalAddresses int64
	db.DB.Model(&models.User{}).Count(&totalUsers)
	db.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&adminCount)
	db.DB.Model(&models.Address{}).Count(&totalAddresses)
	resp.OK(c, gin.H{
		"list": list, "total": total, "page": page, "pageSize": pageSize,
		"stats": gin.H{"totalUsers": totalUsers, "adminCount": adminCount, "totalAddresses": totalAddresses},
	})
}

func authAdminUnbindProduct(c *gin.Context) {
	userID, ok := parseID(c, "userId")
	if !ok {
		resp.Err(c, 400, 400, "参数无效")
		return
	}
	pk := strings.TrimSpace(c.Param("productKey"))
	if pk == "" {
		resp.Err(c, 400, 400, "参数无效")
		return
	}
	res := db.DB.Where("userId = ? AND productKey = ?", userID, pk).Delete(&models.UserProduct{})
	if res.RowsAffected == 0 {
		resp.Err(c, 404, 404, "该用户未绑定此商品")
		return
	}
	resp.OKMsg(c, "已解除绑定")
}

func authAdminDeleteUser(c *gin.Context) {
	me, _ := ctxUser(c)
	userID, ok := parseID(c, "userId")
	if !ok {
		resp.Err(c, 400, 400, "参数无效")
		return
	}
	if me.ID == userID {
		resp.Err(c, 400, 400, "不能删除当前登录用户")
		return
	}
	var target models.User
	if err := db.DB.First(&target, userID).Error; err != nil {
		resp.Err(c, 404, 404, "用户不存在")
		return
	}
	if target.Role == "admin" {
		var ac int64
		db.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&ac)
		if ac <= 1 {
			resp.Err(c, 400, 400, "不能删除最后一个管理员")
			return
		}
	}
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var oids []int
		tx.Model(&models.Order{}).Where("userId = ?", userID).Pluck("id", &oids)
		if len(oids) > 0 {
			tx.Where("orderId IN ?", oids).Delete(&models.OrderLog{})
		}
		tx.Where("userId = ?", userID).Delete(&models.Order{})
		tx.Where("userId = ?", userID).Delete(&models.Address{})
		tx.Where("userId = ?", userID).Delete(&models.Message{})
		tx.Where("userId = ?", userID).Delete(&models.UserProduct{})
		return tx.Delete(&models.User{}, userID).Error
	})
	if err != nil {
		resp.Err(c, 500, 500, err.Error())
		return
	}
	resp.OKMsg(c, "已删除用户")
}

func authMyProducts(c *gin.Context) {
	u, ok := ctxUser(c)
	if !ok {
		return
	}
	var ups []models.UserProduct
	db.DB.Where("userId = ?", u.ID).Order("createdAt DESC").Find(&ups)
	keys := make([]string, 0, len(ups))
	for _, x := range ups {
		keys = append(keys, x.ProductKey)
	}
	if len(keys) == 0 {
		resp.OK(c, []interface{}{})
		return
	}
	var products []models.InventoryProduct
	db.DB.Preload("Category").Where("serialNumber IN ?", keys).Find(&products)

	rawSlugs := make([]string, 0)
	for _, p := range products {
		s := strings.TrimSpace(p.GuideSlug)
		if s != "" {
			rawSlugs = append(rawSlugs, s)
		}
	}
	validSlug := map[string]bool{}
	if len(rawSlugs) > 0 {
		var guides []models.DeviceGuide
		db.DB.Select("slug").Where("slug IN ? AND status = ?", uniqueStr(rawSlugs), "active").Find(&guides)
		for _, g := range guides {
			if g.Slug != nil && *g.Slug != "" {
				validSlug[*g.Slug] = true
			}
		}
	}
	invCatNames := make([]string, 0)
	seen := map[string]bool{}
	for _, p := range products {
		if p.Category != nil && p.Category.Name != "" && !seen[p.Category.Name] {
			seen[p.Category.Name] = true
			invCatNames = append(invCatNames, p.Category.Name)
		}
	}
	invNameToPcID := map[string]int{}
	invNameToNameEn := map[string]string{}
	defaultSlugByPcID := map[int]string{}
	if len(invCatNames) > 0 {
		var pcs []models.ProductCategory
		db.DB.Where("name IN ?", invCatNames).Find(&pcs)
		pcIDs := make([]int, 0)
		for _, pc := range pcs {
			invNameToPcID[pc.Name] = pc.ID
			invNameToNameEn[pc.Name] = pc.NameEn
			pcIDs = append(pcIDs, pc.ID)
		}
		if len(pcIDs) > 0 {
			var dgs []models.DeviceGuide
			db.DB.Where("categoryId IN ? AND status = ?", pcIDs, "active").Order("sortOrder ASC, id ASC").Find(&dgs)
			for _, g := range dgs {
				if g.CategoryID != nil && g.Slug != nil && *g.Slug != "" {
					if _, ok := defaultSlugByPcID[*g.CategoryID]; !ok {
						defaultSlugByPcID[*g.CategoryID] = strings.TrimSpace(*g.Slug)
					}
				}
			}
		}
	}
	resolveSlug := func(p *models.InventoryProduct) string {
		raw := strings.TrimSpace(p.GuideSlug)
		if raw != "" && validSlug[raw] {
			return raw
		}
		cn := ""
		if p.Category != nil {
			cn = p.Category.Name
		}
		if pcID, ok := invNameToPcID[cn]; ok {
			if s, ok := defaultSlugByPcID[pcID]; ok {
				return s
			}
		}
		return ""
	}
	effectiveSlugs := make([]string, 0)
	infoMap := map[string]struct {
		productName, categoryName, categoryNameEn, guideSlug string
	}{}
	for _, p := range products {
		gs := resolveSlug(&p)
		if gs != "" {
			effectiveSlugs = append(effectiveSlugs, gs)
		}
		cn := ""
		if p.Category != nil {
			cn = p.Category.Name
		}
		infoMap[p.SerialNumber] = struct {
			productName, categoryName, categoryNameEn, guideSlug string
		}{p.Name, cn, invNameToNameEn[cn], gs}
	}
	guideBySlug := map[string]struct {
		GuideID, CategoryID int
		IconURL, IconURLThumb, Icon string
	}{}
	if len(effectiveSlugs) > 0 {
		var gds []models.DeviceGuide
		db.DB.Where("slug IN ?", uniqueStr(effectiveSlugs)).Select("id", "slug", "iconUrl", "iconUrlThumb", "icon", "categoryId").Find(&gds)
		for _, g := range gds {
			sk := ""
			if g.Slug != nil {
				sk = *g.Slug
			}
			cid := 0
			if g.CategoryID != nil {
				cid = *g.CategoryID
			}
			guideBySlug[sk] = struct {
				GuideID, CategoryID int
				IconURL, IconURLThumb, Icon string
			}{g.ID, cid, g.IconURL, g.IconURLThumb, g.Icon}
		}
	}
	out := make([]gin.H, 0, len(ups))
	for _, l := range ups {
		info := infoMap[l.ProductKey]
		g := guideBySlug[info.guideSlug]
		out = append(out, gin.H{
			"productKey":     l.ProductKey,
			"productName":    firstStr(info.productName, l.ProductKey),
			"categoryName":   info.categoryName,
			"categoryNameEn": info.categoryNameEn,
			"guideSlug":      info.guideSlug,
			"guideId":        nullInt(g.GuideID),
			"categoryId":     nullInt(g.CategoryID),
			"iconUrl":        g.IconURL,
			"iconUrlThumb":   g.IconURLThumb,
			"guideIcon":      g.Icon,
			"boundAt":        l.CreatedAt,
		})
	}
	resp.OK(c, out)
}

func firstStr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func nullInt(v int) interface{} {
	if v == 0 {
		return nil
	}
	return v
}

func uniqueStr(in []string) []string {
	m := map[string]bool{}
	out := make([]string, 0)
	for _, s := range in {
		if s == "" || m[s] {
			continue
		}
		m[s] = true
		out = append(out, s)
	}
	return out
}

func authBindProduct(c *gin.Context) {
	u, ok := ctxUser(c)
	if !ok {
		return
	}
	var body struct {
		Sn string `json:"sn"`
	}
	_ = c.ShouldBindJSON(&body)
	pk := strings.TrimSpace(body.Sn)
	if pk == "" {
		resp.Err(c, 400, 400, "序列号不能为空")
		return
	}
	var product models.InventoryProduct
	if err := db.DB.Where("serialNumber = ?", pk).First(&product).Error; err != nil {
		resp.Err(c, 404, 404, "未找到该序列号对应的商品")
		return
	}
	if product.Status != "active" {
		resp.Err(c, 400, 400, "商品已下架")
		return
	}
	var existing models.UserProduct
	err := db.DB.Where("productKey = ?", pk).First(&existing).Error
	if err == nil {
		if existing.UserID == u.ID {
			gs := strings.TrimSpace(product.GuideSlug)
			resp.JSON(c, 0, gin.H{"data": gin.H{"productKey": pk, "productName": product.Name, "guideSlug": gs}, "message": "绑定成功"})
			return
		}
		resp.Err(c, 400, 400, "该商品已被其他账号绑定")
		return
	}
	db.DB.Create(&models.UserProduct{UserID: u.ID, ProductKey: pk})
	gs := strings.TrimSpace(product.GuideSlug)
	resp.JSON(c, 0, gin.H{"data": gin.H{"productKey": pk, "productName": product.Name, "guideSlug": gs}, "message": "绑定成功"})
}

func authBindByQrImage(c *gin.Context) {
	u, ok := ctxUser(c)
	if !ok {
		return
	}
	fh, err := c.FormFile("image")
	if err != nil {
		resp.Err(c, 400, 400, "请上传图片")
		return
	}
	f, err := fh.Open()
	if err != nil {
		resp.Err(c, 400, 400, "读取失败")
		return
	}
	defer f.Close()
	buf, err := io.ReadAll(f)
	if err != nil {
		resp.Err(c, 500, 500, "读取失败")
		return
	}
	img, err := imaging.Decode(bytes.NewReader(buf))
	if err != nil {
		resp.Err(c, 400, 400, "无效图片")
		return
	}
	b := img.Bounds()
	if b.Dx() > 1200 || b.Dy() > 1200 {
		img = imaging.Fit(img, 1200, 1200, imaging.Lanczos)
	}
	raw := decodeQRFromImage(img)
	if raw == "" {
		resp.Err(c, 400, 400, "未能识别二维码，请上传清晰的商品二维码图片")
		return
	}
	sn, guide := parseSnGuideFromRaw(raw)
	sn = strings.TrimSpace(sn)
	if sn == "" {
		resp.Err(c, 400, 400, "二维码中未包含序列号，请使用商品绑定二维码")
		return
	}
	var product models.InventoryProduct
	if err := db.DB.Where("serialNumber = ?", sn).First(&product).Error; err != nil {
		resp.Err(c, 404, 404, "未找到该序列号对应的商品")
		return
	}
	if product.Status != "active" {
		resp.Err(c, 400, 400, "商品已下架")
		return
	}
	var existing models.UserProduct
	err = db.DB.Where("productKey = ?", sn).First(&existing).Error
	if err == nil {
		if existing.UserID == u.ID {
			gs := strings.TrimSpace(product.GuideSlug)
			if gs == "" {
				gs = guide
			}
			resp.JSON(c, 0, gin.H{"data": gin.H{"productKey": sn, "productName": product.Name, "guideSlug": gs}, "message": "绑定成功"})
			return
		}
		resp.Err(c, 400, 400, "该商品已被其他账号绑定")
		return
	}
	db.DB.Create(&models.UserProduct{UserID: u.ID, ProductKey: sn})
	gs := strings.TrimSpace(product.GuideSlug)
	if gs == "" {
		gs = guide
	}
	resp.JSON(c, 0, gin.H{"data": gin.H{"productKey": sn, "productName": product.Name, "guideSlug": gs}, "message": "绑定成功"})
}

func decodeQRFromImage(img image.Image) string {
	bmp, err := zxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return ""
	}
	r := qrcode.NewQRCodeReader()
	res, err := r.Decode(bmp, nil)
	if err != nil || res == nil {
		return ""
	}
	return strings.TrimSpace(res.GetText())
}

func parseSnGuideFromRaw(raw string) (sn, guide string) {
	raw = strings.TrimSpace(raw)
	tryURL := raw
	if !strings.HasPrefix(tryURL, "http") {
		tryURL = "http://dummy/" + strings.TrimLeft(tryURL, "/")
	}
	u, err := url.Parse(tryURL)
	if err == nil {
		sn = u.Query().Get("sn")
		guide = u.Query().Get("guide")
		if sn != "" {
			return sn, guide
		}
	}
	reSn := regexp.MustCompile(`[?&]sn=([^&]+)`)
	reG := regexp.MustCompile(`[?&]guide=([^&]+)`)
	if m := reSn.FindStringSubmatch(raw); len(m) > 1 {
		sn, _ = url.QueryUnescape(strings.ReplaceAll(m[1], "+", " "))
	}
	if m := reG.FindStringSubmatch(raw); len(m) > 1 {
		guide, _ = url.QueryUnescape(strings.ReplaceAll(m[1], "+", " "))
	}
	return sn, guide
}

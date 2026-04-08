package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "xyKitchen 服务运行中", "timestamp": time.Now().UTC().Format(time.RFC3339)})
}

func AnalyticsPageView(c *gin.Context) {
	var body struct {
		App  string `json:"app"`
		Path string `json:"path"`
	}
	_ = c.ShouldBindJSON(&body)
	app := body.App
	if app == "" {
		app = "toc"
	}
	if app != "toc" && app != "outlet" && app != "mp" {
		app = "toc"
	}
	p := strings.TrimSpace(body.Path)
	if p == "" {
		p = "/"
	}
	if len(p) > 500 {
		p = p[:500]
	}
	pageKey := app + ":" + p
	visitDate := time.Now().Format("2006-01-02")
	err := db.DB.Exec(
		"INSERT INTO page_visit_daily (page_key, visit_date, `count`, created_at, updated_at) VALUES (?, ?, 1, NOW(), NOW()) ON DUPLICATE KEY UPDATE `count` = `count` + 1, updated_at = NOW()",
		pageKey, visitDate,
	).Error
	if err != nil {
		resp.Err(c, 500, 1, "记录失败")
		return
	}
	resp.OKMsg(c, "ok")
}

func AnalyticsStats(c *gin.Context) {
	type row struct {
		PageKey string  `gorm:"column:pageKey"`
		Total   float64 `gorm:"column:total"`
		D7      float64 `gorm:"column:d7"`
		D30     float64 `gorm:"column:d30"`
		D90     float64 `gorm:"column:d90"`
	}
	var rows []row
	sql := `SELECT page_key AS pageKey,
		SUM(` + "`count`" + `) AS total,
		SUM(CASE WHEN visit_date >= DATE_SUB(CURDATE(), INTERVAL 6 DAY) THEN ` + "`count`" + ` ELSE 0 END) AS d7,
		SUM(CASE WHEN visit_date >= DATE_SUB(CURDATE(), INTERVAL 29 DAY) THEN ` + "`count`" + ` ELSE 0 END) AS d30,
		SUM(CASE WHEN visit_date >= DATE_SUB(CURDATE(), INTERVAL 89 DAY) THEN ` + "`count`" + ` ELSE 0 END) AS d90
		FROM page_visit_daily
		GROUP BY page_key
		HAVING SUM(` + "`count`" + `) > 0
		ORDER BY total DESC`
	if err := db.DB.Raw(sql).Scan(&rows).Error; err != nil {
		resp.Err(c, 500, 1, err.Error())
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"pageKey": r.PageKey, "total": int(r.Total), "last7Days": int(r.D7), "last30Days": int(r.D30), "lastQuarter": int(r.D90),
		})
	}
	resp.OK(c, gin.H{"rows": out})
}

func SeedData(c *gin.Context) {
	users := queryInt(c, "users", 10000)
	products := queryInt(c, "products", 10000)
	var cats []models.InventoryCategory
	db.DB.Find(&cats)
	if len(cats) == 0 {
		for _, name := range []string{"空调", "冰箱", "洗衣机", "热水器", "其他"} {
			db.DB.Create(&models.InventoryCategory{Name: name, SortOrder: 0, Status: "active"})
		}
		db.DB.Find(&cats)
	}
	catIDs := make([]int, len(cats))
	for i, c := range cats {
		catIDs[i] = c.ID
	}
	const batch = 500
	dummyHash := "$2b$10$dummyHashForSeedDataOnly00000000000000000000000000"
	names1 := []string{"张", "李", "王", "刘", "陈"}
	names2 := []string{"伟", "芳", "娜", "敏", "静"}
	prefixes := []string{"空调", "冰箱", "洗衣机"}
	suffixes := []string{"Pro", "Max", "Air"}
	tagsPool := []string{"新品", "热销", "特价"}
	for i := 0; i < users; i += batch {
		n := batch
		if i+n > users {
			n = users - i
		}
		for j := 0; j < n; j++ {
			idx := i + j
			b := make([]byte, 4)
			_, _ = rand.Read(b)
			hexv := hex.EncodeToString(b)
			phone := fmt.Sprintf("1%02d%08d", 30+idx%70, idx%100000000)
			email := fmt.Sprintf("test%d_%s@seed.local", idx, hexv)
			uname := fmt.Sprintf("test_%s_%d", hexv, idx)
			nick := names1[idx%len(names1)] + names2[idx%len(names2)] + fmt.Sprintf("%d", idx%999)
			em := email
			u := models.User{Username: uname, Password: dummyHash, Email: &em, Phone: phone, Nickname: nick, Role: "user", Status: "active"}
			_ = db.DB.Create(&u).Error
		}
	}
	rint := func(idx int) int { return idx % 100000 }
	for i := 0; i < products; i += batch {
		n := batch
		if i+n > products {
			n = products - i
		}
		for j := 0; j < n; j++ {
			idx := i + j
			b := make([]byte, 3)
			_, _ = rand.Read(b)
			hexv := strings.ToUpper(hex.EncodeToString(b))
			tags := ""
			if rint(idx)%4 != 0 {
				tags = tagsPool[idx%len(tagsPool)]
			}
			st := "active"
			if rint(idx)%10 == 0 {
				st = "inactive"
			}
			p := models.InventoryProduct{
				CategoryID:   catIDs[idx%len(catIDs)],
				Name:         prefixes[idx%len(prefixes)] + " " + suffixes[idx%len(suffixes)] + "-" + hexv,
				SerialNumber: fmt.Sprintf("SN%06d%s", idx, hexv),
				SortOrder:    rint(idx) % 100,
				Status:       st,
				Tags:         tags,
			}
			_ = db.DB.Create(&p).Error
		}
	}
	resp.OKMsg(c, fmt.Sprintf("已生成 %d 个用户、%d 个库存商品", users, products))
}

func MediaCosStream(c *gin.Context) {
	key := strings.TrimSpace(c.Query("key"))
	if key == "" {
		resp.Err(c, 400, 400, "缺少 key")
		return
	}
	decoded, err := url.QueryUnescape(key)
	if err != nil {
		c.Status(400)
		return
	}
	if !services.IsKeyAllowedForProxy(decoded) {
		resp.Err(c, 400, 400, "非法 key")
		return
	}
	_ = services.StreamCosObjectToResponse(c.Request.Context(), decoded, c.Writer)
}

func AdminGenerateThumbs(c *gin.Context, cfg *config.Config) {
	_ = cfg
	processed, failed, skipped := 0, 0, 0
	var guides []models.DeviceGuide
	db.DB.Select("id", "iconUrl", "coverImage", "qrcodeUrl").Find(&guides)
	seen := map[string]bool{}
	for _, g := range guides {
		for _, u := range []string{g.IconURL, g.CoverImage, g.QrcodeURL} {
			if u != "" && services.IsCosUploadURL(u) && !seen[u] {
				seen[u] = true
			}
		}
	}
	var hcs []models.HomeConfig
	db.DB.Where("imageUrl != ?", "").Select("imageUrl").Find(&hcs)
	for _, h := range hcs {
		if h.ImageURL != "" && services.IsCosUploadURL(h.ImageURL) && !seen[h.ImageURL] {
			seen[h.ImageURL] = true
		}
	}
	ctx := context.Background()
	for u := range seen {
		fname := filenameFromCosURL(u)
		if fname == "" {
			skipped++
			continue
		}
		buf, err := fetchImageBuffer(ctx, u)
		if err != nil || len(buf) == 0 {
			skipped++
			continue
		}
		tb, tct := services.GenerateThumbBuffer(buf, "image/jpeg")
		if len(tb) == 0 {
			skipped++
			continue
		}
		if _, err := services.UploadThumb(ctx, tb, fname, tct); err != nil {
			failed++
			continue
		}
		processed++
	}
	resp.OK(c, gin.H{
		"message":   fmt.Sprintf("已处理 %d 张缩略图，失败 %d，跳过 %d", processed, failed, skipped),
		"processed": processed, "failed": failed, "skipped": skipped,
	})
}

func fetchImageBuffer(ctx context.Context, u string) ([]byte, error) {
	key := services.URLToKey(u)
	if key != "" {
		return services.GetObjectBuffer(ctx, key)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Vino-Backend/1.0")
	respHTTP, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer respHTTP.Body.Close()
	if respHTTP.StatusCode != 200 {
		return nil, fmt.Errorf("http %d", respHTTP.StatusCode)
	}
	return io.ReadAll(respHTTP.Body)
}

func filenameFromCosURL(u string) string {
	base := services.CosBase() + "/vino/uploads/"
	if !strings.HasPrefix(u, base) {
		return ""
	}
	path := strings.TrimPrefix(u, base)
	if path == "" || strings.Contains(path, "thumb/") {
		return ""
	}
	return path
}

func I18nList(c *gin.Context) {
	var items []models.I18nText
	db.DB.Order("`key` ASC").Find(&items)
	resp.OK(c, items)
}

func I18nBulkUpsert(c *gin.Context) {
	var body struct {
		Rows []struct {
			Key string `json:"key"`
			Zh  string `json:"zh"`
			En  string `json:"en"`
		} `json:"rows"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Err(c, 400, 1, "rows required")
		return
	}
	for _, r := range body.Rows {
		if r.Key == "" {
			continue
		}
		var it models.I18nText
		err := db.DB.Where("`key` = ?", r.Key).First(&it).Error
		if err != nil {
			db.DB.Create(&models.I18nText{Key: r.Key, Zh: r.Zh, En: r.En})
			continue
		}
		upd := false
		if r.Zh != "" {
			it.Zh = r.Zh
			upd = true
		}
		if r.En != "" {
			it.En = r.En
			upd = true
		}
		if upd {
			db.DB.Save(&it)
		}
	}
	var items []models.I18nText
	db.DB.Order("`key` ASC").Find(&items)
	resp.OK(c, items)
}

func I18nUpdate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var it models.I18nText
	if err := db.DB.First(&it, id).Error; err != nil {
		resp.Err(c, 404, 1, "not found")
		return
	}
	var body struct {
		Zh, En string
	}
	_ = c.ShouldBindJSON(&body)
	if body.Zh != "" {
		it.Zh = body.Zh
	}
	if body.En != "" {
		it.En = body.En
	}
	db.DB.Save(&it)
	resp.OK(c, it)
}

func I18nRemove(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	db.DB.Delete(&models.I18nText{}, id)
	resp.OKMsg(c, "deleted")
}

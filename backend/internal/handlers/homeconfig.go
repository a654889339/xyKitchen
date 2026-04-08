package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/url"
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

func fixHomeProxyURL(u string) string {
	if u == "" {
		return u
	}
	s := strings.TrimSpace(u)
	prefix := "/api/media/cos?key="
	if strings.HasPrefix(s, prefix) {
		key, err := url.QueryUnescape(s[len(prefix):])
		if err != nil {
			return s
		}
		return services.CosBase() + "/" + key
	}
	return s
}

func hcList(c *gin.Context) {
	q := db.DB.Model(&models.HomeConfig{})
	if sec := c.Query("section"); sec != "" {
		q = q.Where("section = ?", sec)
	}
	if c.Query("all") == "" {
		q = q.Where("status = ?", "active")
	}
	var items []models.HomeConfig
	q.Order("section ASC, sortOrder ASC, id ASC").Find(&items)
	out := make([]gin.H, 0, len(items))
	for _, it := range items {
		raw, _ := json.Marshal(it)
		var o gin.H
		_ = json.Unmarshal(raw, &o)
		thumb := ""
		if x, ok := o["imageUrlThumb"].(string); ok {
			thumb = x
		}
		o["imageUrl"] = fixHomeProxyURL(it.ImageURL)
		o["imageUrlThumb"] = fixHomeProxyURL(strings.TrimSpace(thumb))
		out = append(out, o)
	}
	resp.OK(c, out)
}

func hcCreate(c *gin.Context) {
	var body models.HomeConfig
	if err := c.ShouldBindJSON(&body); err != nil {
		resp.Err(c, 500, 1, err.Error())
		return
	}
	if err := db.DB.Create(&body).Error; err != nil {
		resp.Err(c, 500, 1, err.Error())
		return
	}
	resp.OK(c, body)
}

func hcUpdate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item models.HomeConfig
	if err := db.DB.First(&item, id).Error; err != nil {
		resp.Err(c, 404, 1, "配置不存在")
		return
	}
	var patch map[string]interface{}
	if err := c.ShouldBindJSON(&patch); err != nil {
		resp.Err(c, 500, 1, err.Error())
		return
	}
	raw, _ := json.Marshal(patch)
	_ = json.Unmarshal(raw, &item)
	item.ID = id
	if err := db.DB.Save(&item).Error; err != nil {
		resp.Err(c, 500, 1, err.Error())
		return
	}
	resp.OK(c, item)
}

func hcRemove(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := db.DB.Delete(&models.HomeConfig{}, id).Error; err != nil {
		resp.Err(c, 500, 1, err.Error())
		return
	}
	resp.OKMsg(c, "删除成功")
}

func hcUploadImage(c *gin.Context, cfg *config.Config) {
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
		ext = ".png"
	}
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	filename := "homeconfig-" + strconv.FormatInt(time.Now().UnixMilli(), 10) + "-" + hex.EncodeToString(b) + ext
	ct := fh.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/png"
	}
	urlu, thumb, err := services.UploadWithThumb(c.Request.Context(), buf, filename, ct, 0)
	if err != nil {
		resp.Err(c, 500, 1, err.Error())
		return
	}
	resp.OK(c, gin.H{"url": urlu, "thumbUrl": thumb})
}

package services

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/disintegration/imaging"
	cos "github.com/tencentyun/cos-go-sdk-v5"
	_ "golang.org/x/image/webp"
)

const (
	cosBucket = "itsyourturnmy-1256887166"
	cosRegion = "ap-singapore"
)

var cosBaseURL = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cosBucket, cosRegion)

func cosClient() (*cos.Client, error) {
	sid := os.Getenv("COS_SECRET_ID")
	sk := os.Getenv("COS_SECRET_KEY")
	if sid == "" || sk == "" {
		return nil, fmt.Errorf("COS not configured")
	}
	u, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cosBucket, cosRegion))
	b := &cos.BaseURL{BucketURL: u}
	return cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  sid,
			SecretKey: sk,
		},
	}), nil
}

func CosBase() string { return cosBaseURL }

// URLToKey 从完整 COS URL 解析 object key（与 Node urlToKey 一致）
func URLToKey(fullURL string) string {
	if fullURL == "" {
		return ""
	}
	u := strings.TrimSpace(fullURL)
	if !strings.HasPrefix(u, cosBaseURL+"/") {
		return ""
	}
	q := strings.Index(u, "?")
	if q >= 0 {
		u = u[:q]
	}
	key := strings.TrimPrefix(u, cosBaseURL+"/")
	if key == "" {
		return ""
	}
	return key
}

// GetThumbURL 由原图 URL 推导缩略图 URL（仅本站 COS 直链）
// IsCosUploadURL 与 Node isCosUploadUrl 一致：本桶 vino/uploads 原图（非 thumb 子目录）
func IsCosUploadURL(u string) bool {
	if u == "" {
		return false
	}
	base := cosBaseURL + "/vino/uploads/"
	if !strings.HasPrefix(u, base) {
		return false
	}
	return !strings.Contains(u, "/vino/uploads/thumb/")
}

func GetThumbURL(originalURL string) string {
	if originalURL == "" {
		return ""
	}
	base := cosBaseURL + "/vino/uploads/"
	if !strings.HasPrefix(originalURL, base) {
		return ""
	}
	suffix := strings.TrimPrefix(originalURL, base)
	if suffix == "" || strings.Contains(suffix, "thumb/") {
		return ""
	}
	return cosBaseURL + "/vino/uploads/thumb/" + suffix
}

const thumbMaxWidth = 400

// GenerateThumbBuffer 生成缩略图（默认最大边 400px，与 Node 一致）
func GenerateThumbBuffer(buf []byte, contentType string) ([]byte, string) {
	return generateThumbBufferMax(buf, contentType, thumbMaxWidth)
}

func generateThumbBufferMax(buf []byte, contentType string, maxW int) ([]byte, string) {
	if maxW <= 0 {
		maxW = thumbMaxWidth
	}
	img, err := imaging.Decode(bytes.NewReader(buf))
	if err != nil {
		return nil, ""
	}
	if img.Bounds().Dx() > maxW || img.Bounds().Dy() > maxW {
		img = imaging.Fit(img, maxW, maxW, imaging.Lanczos)
	}
	lower := strings.ToLower(contentType)
	var out bytes.Buffer
	switch {
	case strings.Contains(lower, "png"):
		_ = png.Encode(&out, img)
		return out.Bytes(), "image/png"
	case strings.Contains(lower, "webp"):
		_ = jpeg.Encode(&out, img, &jpeg.Options{Quality: 82})
		return out.Bytes(), "image/jpeg"
	default:
		_ = jpeg.Encode(&out, img, &jpeg.Options{Quality: 82})
		return out.Bytes(), "image/jpeg"
	}
}

// UploadThumb 上传到 vino/uploads/thumb/
func UploadThumb(ctx context.Context, buf []byte, filename, contentType string) (string, error) {
	c, err := cosClient()
	if err != nil {
		return "", err
	}
	key := "vino/uploads/thumb/" + filename
	_, err = c.Object.Put(ctx, key, bytes.NewReader(buf), &cos.ObjectPutOptions{
		ACLHeaderOptions: &cos.ACLHeaderOptions{
			XCosACL: "public-read",
		},
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	})
	if err != nil {
		return "", err
	}
	return cosBaseURL + "/" + key, nil
}

// UploadWithThumb 上传原图并生成缩略图（maxWidth 0 使用默认 400）
func UploadWithThumb(ctx context.Context, buf []byte, filename, contentType string, maxWidth int) (url string, thumbURL string, err error) {
	url, err = UploadCOS(ctx, buf, filename, contentType)
	if err != nil {
		return "", "", err
	}
	mw := thumbMaxWidth
	if maxWidth > 0 {
		mw = maxWidth
	}
	tb, tct := generateThumbBufferMax(buf, contentType, mw)
	if len(tb) == 0 {
		return url, "", nil
	}
	tu, err := UploadThumb(ctx, tb, filename, tct)
	if err != nil {
		return url, "", nil
	}
	return url, tu, nil
}

func UploadCOS(ctx context.Context, buf []byte, filename, contentType string) (string, error) {
	c, err := cosClient()
	if err != nil {
		return "", err
	}
	key := "vino/uploads/" + filename
	_, err = c.Object.Put(ctx, key, bytes.NewReader(buf), &cos.ObjectPutOptions{
		ACLHeaderOptions: &cos.ACLHeaderOptions{
			XCosACL: "public-read",
		},
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	})
	if err != nil {
		return "", err
	}
	return cosBaseURL + "/" + key, nil
}

func UploadCOSReader(ctx context.Context, r io.Reader, filename, contentType string) (string, error) {
	c, err := cosClient()
	if err != nil {
		return "", err
	}
	key := "vino/uploads/" + filename
	_, err = c.Object.Put(ctx, key, r, &cos.ObjectPutOptions{
		ACLHeaderOptions: &cos.ACLHeaderOptions{
			XCosACL: "public-read",
		},
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	})
	if err != nil {
		return "", err
	}
	return cosBaseURL + "/" + key, nil
}

func IsKeyAllowedForProxy(key string) bool {
	if key == "" || strings.Contains(key, "..") || strings.Contains(key, "\\") {
		return false
	}
	return strings.HasPrefix(key, "vino/uploads/")
}

func StreamCosObjectToResponse(ctx context.Context, key string, w http.ResponseWriter) error {
	c, err := cosClient()
	if err != nil {
		http.Error(w, `{"code":503,"message":"COS 未配置"}`, 503)
		return err
	}
	resp, err := c.Object.Get(ctx, key, nil)
	if err != nil {
		http.Error(w, "", 404)
		return err
	}
	defer resp.Body.Close()
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "public, max-age=300")
	_, err = io.Copy(w, resp.Body)
	return err
}

func GetObjectBuffer(ctx context.Context, key string) ([]byte, error) {
	c, err := cosClient()
	if err != nil {
		return nil, err
	}
	resp, err := c.Object.Get(ctx, key, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

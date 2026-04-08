package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port    int
	NodeEnv string
	JWT     struct {
		Secret    string
		ExpiresIn string
	}
	DB struct {
		Host     string
		Port     int
		Name     string
		User     string
		Password string
	}
	Wechat struct {
		AppID     string
		AppSecret string
	}
	Alipay struct {
		AppID       string
		PrivateKey  string
		PublicKey   string
	}
	Email struct {
		Enabled  bool
		Host     string
		Port     int
		Username string
		Password string
		From     string
		FromName string
	}
	SMS struct {
		Enabled           bool
		SecretID          string
		SecretKey         string
		SmsSdkAppID       string
		SignName          string
		TemplateID        string
		CodeExpireMinutes int
	}
	FrontendURL string
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func atoi(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func Load() *Config {
	c := &Config{}
	c.Port = atoi(getenv("PORT", "5402"), 5402)
	c.NodeEnv = getenv("NODE_ENV", "development")
	c.JWT.Secret = getenv("JWT_SECRET", "xykitchen_default_secret")
	c.JWT.ExpiresIn = getenv("JWT_EXPIRES_IN", "168h")
	c.DB.Host = getenv("DB_HOST", "localhost")
	c.DB.Port = atoi(getenv("DB_PORT", "3311"), 3311)
	c.DB.Name = getenv("DB_NAME", "xykitchen_db")
	c.DB.User = getenv("DB_USER", "root")
	c.DB.Password = getenv("DB_PASSWORD", "xykitchen_secret_2026")
	c.Wechat.AppID = getenv("WECHAT_APPID", "")
	c.Wechat.AppSecret = getenv("WECHAT_SECRET", "")
	c.Alipay.AppID = getenv("ALIPAY_APPID", "")
	c.Alipay.PrivateKey = strings.ReplaceAll(getenv("ALIPAY_PRIVATE_KEY", ""), "\\n", "\n")
	c.Alipay.PublicKey = strings.ReplaceAll(getenv("ALIPAY_PUBLIC_KEY", ""), "\\n", "\n")
	c.Email.Enabled = getenv("EMAIL_ENABLED", "false") == "true"
	c.Email.Host = getenv("EMAIL_HOST", "smtp.qq.com")
	c.Email.Port = atoi(getenv("EMAIL_PORT", "587"), 587)
	c.Email.Username = getenv("EMAIL_USERNAME", "")
	c.Email.Password = getenv("EMAIL_PASSWORD", "")
	c.Email.From = getenv("EMAIL_FROM", "")
	c.Email.FromName = getenv("EMAIL_FROM_NAME", "xyKitchen")
	c.SMS.Enabled = getenv("SMS_ENABLED", "false") == "true"
	c.SMS.SecretID = getenv("TENCENT_SMS_SECRET_ID", "")
	c.SMS.SecretKey = getenv("TENCENT_SMS_SECRET_KEY", "")
	c.SMS.SmsSdkAppID = getenv("TENCENT_SMS_APP_ID", "")
	c.SMS.SignName = getenv("TENCENT_SMS_SIGN_NAME", "xyKitchen")
	c.SMS.TemplateID = getenv("TENCENT_SMS_TEMPLATE_ID", "")
	c.SMS.CodeExpireMinutes = atoi(getenv("SMS_CODE_EXPIRE_MINUTES", "5"), 5)
	c.FrontendURL = getenv("FRONTEND_URL", "http://106.54.50.88:5402")
	return c
}

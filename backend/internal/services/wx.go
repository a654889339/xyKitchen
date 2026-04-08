package services

import (
	"encoding/json"
	"fmt"
	"net/http"

	"xykitchen/backend/internal/config"
)

func WxCode2Session(cfg *config.Config, code string) (openid string, err error) {
	if cfg.Wechat.AppID == "" || cfg.Wechat.AppSecret == "" {
		return "", fmt.Errorf("微信配置缺失")
	}
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		cfg.Wechat.AppID, cfg.Wechat.AppSecret, code)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct {
		Errcode  int    `json:"errcode"`
		Errmsg   string `json:"errmsg"`
		Openid   string `json:"openid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Errcode != 0 {
		return "", fmt.Errorf("%s", out.Errmsg)
	}
	if out.Openid == "" {
		return "", fmt.Errorf("获取openid失败")
	}
	return out.Openid, nil
}

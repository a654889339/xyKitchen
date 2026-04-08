package services

import (
	"fmt"

	"xykitchen/backend/internal/config"

	"github.com/smartwalle/alipay/v3"
)

func AlipayOAuthToken(cfg *config.Config, code string) (userID string, err error) {
	if cfg.Alipay.AppID == "" || cfg.Alipay.PrivateKey == "" {
		return "", fmt.Errorf("支付宝配置缺失")
	}
	isProd := cfg.NodeEnv == "production"
	client, err := alipay.New(cfg.Alipay.AppID, cfg.Alipay.PrivateKey, isProd)
	if err != nil {
		return "", err
	}
	if cfg.Alipay.PublicKey != "" {
		if err := client.LoadAliPayPublicKey(cfg.Alipay.PublicKey); err != nil {
			return "", err
		}
	}
	var p alipay.SystemOauthToken
	p.GrantType = "authorization_code"
	p.Code = code
	r, err := client.SystemOauthToken(p)
	if err != nil {
		return "", err
	}
	if r == nil {
		return "", fmt.Errorf("empty oauth response")
	}
	if r.UserId != "" {
		return r.UserId, nil
	}
	if r.OpenId != "" {
		return r.OpenId, nil
	}
	return "", fmt.Errorf("empty user id")
}

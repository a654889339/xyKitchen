package services

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"xykitchen/backend/internal/config"
)

func loadWechatPayPrivateKey() *rsa.PrivateKey {
	raw := os.Getenv("WECHAT_PAY_PRIVATE_KEY")
	if raw == "" {
		return nil
	}
	if !strings.Contains(raw, "BEGIN") {
		raw = strings.ReplaceAll(raw, "\\n", "\n")
	}
	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil
	}
	k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		k2, err2 := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err2 != nil {
			return nil
		}
		return k2
	}
	if pk, ok := k.(*rsa.PrivateKey); ok {
		return pk
	}
	return nil
}

func IsWechatPayConfigured(cfg *config.Config) bool {
	return cfg.Wechat.AppID != "" &&
		os.Getenv("WECHAT_PAY_MCH_ID") != "" &&
		os.Getenv("WECHAT_PAY_SERIAL_NO") != "" &&
		len(os.Getenv("WECHAT_PAY_API_V3_KEY")) == 32 &&
		loadWechatPayPrivateKey() != nil &&
		os.Getenv("WECHAT_PAY_NOTIFY_URL") != ""
}

func buildPayAuth(method, urlPath, bodyStr string, pk *rsa.PrivateKey) (string, error) {
	mchid := os.Getenv("WECHAT_PAY_MCH_ID")
	serial := os.Getenv("WECHAT_PAY_SERIAL_NO")
	ts := fmt.Sprintf("%d", time.Now().Unix())
	nonce := make([]byte, 16)
	_, _ = rand.Read(nonce)
	nonceStr := fmt.Sprintf("%x", nonce)
	msg := method + "\n" + urlPath + "\n" + ts + "\n" + nonceStr + "\n" + bodyStr + "\n"
	h := sha256.Sum256([]byte(msg))
	sig, err := rsa.SignPKCS1v15(nil, pk, crypto.SHA256, h[:])
	if err != nil {
		return "", err
	}
	signB64 := base64.StdEncoding.EncodeToString(sig)
	return fmt.Sprintf(`WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",timestamp="%s",serial_no="%s",signature="%s"`,
		mchid, nonceStr, ts, serial, signB64), nil
}

func JsapiPrepay(cfg *config.Config, outTradeNo, description string, totalFen int, openid string) (map[string]interface{}, error) {
	pk := loadWechatPayPrivateKey()
	if pk == nil {
		return nil, fmt.Errorf("no private key")
	}
	urlPath := "/v3/pay/transactions/jsapi"
	body := map[string]interface{}{
		"appid":        cfg.Wechat.AppID,
		"mchid":        os.Getenv("WECHAT_PAY_MCH_ID"),
		"description":  description,
		"out_trade_no": outTradeNo,
		"notify_url":   os.Getenv("WECHAT_PAY_NOTIFY_URL"),
		"amount":       map[string]interface{}{"total": totalFen, "currency": "CNY"},
		"payer":        map[string]interface{}{"openid": openid},
	}
	bodyStr, _ := json.Marshal(body)
	auth, err := buildPayAuth("POST", urlPath, string(bodyStr), pk)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest(http.MethodPost, "https://api.mch.weixin.qq.com"+urlPath, bytes.NewReader(bodyStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if resp.StatusCode >= 400 {
		return out, fmt.Errorf("%v", out)
	}
	return out, nil
}

func BuildMiniProgramPayParams(cfg *config.Config, prepayID string) (map[string]interface{}, error) {
	pk := loadWechatPayPrivateKey()
	if pk == nil {
		return nil, fmt.Errorf("no private key")
	}
	appID := cfg.Wechat.AppID
	ts := fmt.Sprintf("%d", time.Now().Unix())
	nonce := make([]byte, 16)
	_, _ = rand.Read(nonce)
	nonceStr := fmt.Sprintf("%x", nonce)
	pkg := "prepay_id=" + prepayID
	signStr := appID + "\n" + ts + "\n" + nonceStr + "\n" + pkg + "\n"
	h := sha256.Sum256([]byte(signStr))
	sig, err := rsa.SignPKCS1v15(nil, pk, crypto.SHA256, h[:])
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"appId":     appID,
		"timeStamp": ts,
		"nonceStr":  nonceStr,
		"package":   pkg,
		"signType":  "RSA",
		"paySign":   base64.StdEncoding.EncodeToString(sig),
	}, nil
}

// DecryptNotifyResource 解密微信支付 APIv3 回调 resource（AES-256-GCM）
func DecryptNotifyResource(res map[string]interface{}) (map[string]interface{}, error) {
	ciphertext, _ := res["ciphertext"].(string)
	ad, _ := res["associated_data"].(string)
	nonceStr, _ := res["nonce"].(string)
	keyStr := os.Getenv("WECHAT_PAY_API_V3_KEY")
	if len(keyStr) != 32 {
		return nil, fmt.Errorf("invalid api v3 key")
	}
	key := []byte(keyStr)
	nonceB := []byte(nonceStr)
	buf, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plain, err := gcm.Open(nil, nonceB, buf, []byte(ad))
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(plain, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func init() {
	_ = io.EOF
}

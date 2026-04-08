package services

import (
	"fmt"
	"strings"

	"xykitchen/backend/internal/config"

	mail "github.com/go-mail/mail/v2"
)

func SendEmailCode(cfg *config.Config, to string) error {
	if !cfg.Email.Enabled {
		return fmt.Errorf("邮件服务未启用")
	}
	if err := EmailCanSend(to); err != nil {
		return err
	}
	code := genEmailCode()
	d := mail.NewDialer(cfg.Email.Host, cfg.Email.Port, cfg.Email.Username, cfg.Email.Password)
	d.StartTLSPolicy = mail.OpportunisticStartTLS
	if cfg.Email.Port == 465 {
		d.SSL = true
	}
	m := mail.NewMessage()
	from := cfg.Email.From
	if cfg.Email.FromName != "" {
		from = fmt.Sprintf("%q <%s>", cfg.Email.FromName, cfg.Email.From)
	}
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Vino服务 - 邮箱验证码")
	body := fmt.Sprintf(`<div style="max-width:480px;margin:0 auto;font-family:sans-serif;">
<div style="background:#B91C1C;color:#fff;padding:20px 24px;border-radius:8px 8px 0 0;"><h2 style="margin:0;">Vino 服务站</h2></div>
<div style="border:1px solid #eee;border-top:none;padding:24px;border-radius:0 0 8px 8px;">
<p>您好，您的邮箱验证码为：</p>
<div style="text-align:center;margin:24px 0;"><span style="font-size:32px;font-weight:bold;letter-spacing:8px;color:#B91C1C;">%s</span></div>
<p style="color:#666;font-size:13px;">验证码 5 分钟内有效，请勿泄露给他人。</p>
</div></div>`, code)
	m.SetBody("text/html", body)
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	EmailSetCode(strings.ToLower(strings.TrimSpace(to)), code)
	return nil
}

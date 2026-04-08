package services

import (
	"crypto/rand"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// NormalizePhone 与 Node smsService.normalizePhone 一致
func NormalizePhone(phone string) string {
	p := strings.ReplaceAll(phone, " ", "")
	p = strings.ReplaceAll(p, "\t", "")
	if len(p) == 11 && p[0] == '1' {
		return p
	}
	if strings.HasPrefix(p, "+86") {
		p = strings.TrimPrefix(p, "+86")
		p = strings.TrimLeft(p, " ")
	}
	return p
}

type emailRec struct {
	code        string
	expiresAt   int64
	attempts    int
	used        bool
	lockedUntil int64
}

type emailSendRec struct {
	times    []int64
	lastSend int64
}

type smsRec struct {
	code      string
	expiresAt int64
	used      bool
}

var (
	emailMu    sync.Mutex
	emailCodes = map[string]*emailRec{}
	emailSendM = map[string]*emailSendRec{}

	smsMu    sync.Mutex
	smsCodes = map[string]*smsRec{}
	smsSendM = map[string]int64{}
)

const (
	emailCodeExpire   = 5 * time.Minute
	emailCooldown     = 60 * time.Second
	emailMaxPerHour   = 5
	emailMaxAttempts  = 5
	emailLockDuration = 30 * time.Minute
	smsCooldown       = 60 * time.Second
)

func genEmailCode() string {
	b := make([]byte, 3)
	_, _ = io.ReadFull(rand.Reader, b)
	n := int(b[0])<<16 | int(b[1])<<8 | int(b[2])
	if n < 0 {
		n = -n
	}
	return fmt.Sprintf("%06d", n%1000000)
}

func genSMSCode() string {
	b := make([]byte, 2)
	_, _ = rand.Read(b)
	n := int(b[0])<<8 | int(b[1])
	if n < 0 {
		n = -n
	}
	return fmt.Sprintf("%06d", 100000+n%900000)
}

func EmailCanSend(email string) error {
	emailMu.Lock()
	defer emailMu.Unlock()
	r := emailSendM[email]
	if r == nil {
		return nil
	}
	now := time.Now()
	if r.lastSend > 0 && now.Sub(time.Unix(0, r.lastSend)) < emailCooldown {
		wait := int((emailCooldown - now.Sub(time.Unix(0, r.lastSend))).Seconds()) + 1
		return fmt.Errorf("发送过于频繁，请 %d 秒后再试", wait)
	}
	cut := now.Add(-time.Hour).UnixNano()
	var kept []int64
	for _, t := range r.times {
		if t > cut {
			kept = append(kept, t)
		}
	}
	r.times = kept
	if len(r.times) >= emailMaxPerHour {
		return fmt.Errorf("已超过每小时发送上限，请稍后再试")
	}
	return nil
}

func EmailSetCode(email, code string) {
	emailMu.Lock()
	defer emailMu.Unlock()
	emailCodes[email] = &emailRec{
		code:      code,
		expiresAt: time.Now().Add(emailCodeExpire).UnixNano(),
		used:      false,
	}
	rec := emailSendM[email]
	if rec == nil {
		rec = &emailSendRec{}
		emailSendM[email] = rec
	}
	rec.times = append(rec.times, time.Now().UnixNano())
	rec.lastSend = time.Now().UnixNano()
}

func EmailVerify(email, code string) (bool, string) {
	emailMu.Lock()
	defer emailMu.Unlock()
	vc := emailCodes[email]
	if vc == nil {
		return false, "请先获取验证码"
	}
	if vc.lockedUntil > time.Now().UnixNano() {
		return false, "验证次数过多，请稍后再试"
	}
	if vc.attempts >= emailMaxAttempts {
		vc.lockedUntil = time.Now().Add(emailLockDuration).UnixNano()
		vc.attempts = 0
		return false, "验证次数过多，已锁定 30 分钟"
	}
	vc.attempts++
	if vc.used {
		return false, "验证码已使用，请重新获取"
	}
	if time.Now().UnixNano() > vc.expiresAt {
		delete(emailCodes, email)
		return false, "验证码已过期，请重新获取"
	}
	if vc.code != code {
		return false, "验证码错误"
	}
	vc.used = true
	return true, ""
}

func SMSCanSend(phone string) error {
	smsMu.Lock()
	defer smsMu.Unlock()
	key := NormalizePhone(phone)
	last := smsSendM[key]
	if last == 0 {
		return nil
	}
	if time.Since(time.Unix(0, last)) < smsCooldown {
		wait := int((smsCooldown - time.Since(time.Unix(0, last))).Seconds()) + 1
		return fmt.Errorf("发送过于频繁，请 %d 秒后再试", wait)
	}
	return nil
}

func SMSSetCode(phone, code string, expire time.Duration) {
	smsMu.Lock()
	defer smsMu.Unlock()
	key := NormalizePhone(phone)
	smsCodes[key] = &smsRec{
		code:      code,
		expiresAt: time.Now().Add(expire).UnixNano(),
		used:      false,
	}
	smsSendM[key] = time.Now().UnixNano()
}

func SMSVerify(phone, code string) (bool, string) {
	smsMu.Lock()
	defer smsMu.Unlock()
	key := NormalizePhone(phone)
	vc := smsCodes[key]
	if vc == nil {
		return false, "请先获取验证码"
	}
	if vc.used {
		return false, "验证码已使用，请重新获取"
	}
	if time.Now().UnixNano() > vc.expiresAt {
		delete(smsCodes, key)
		return false, "验证码已过期，请重新获取"
	}
	if vc.code != code {
		return false, "验证码错误"
	}
	vc.used = true
	return true, ""
}

package services

import (
	"context"
	"fmt"
	"time"

	"xykitchen/backend/internal/config"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

func SendSMSCode(cfg *config.Config, phone string) error {
	if !cfg.SMS.Enabled {
		return fmt.Errorf("短信服务未启用，请配置 TENCENT_SMS_* 环境变量并设置 SMS_ENABLED=true")
	}
	key := NormalizePhone(phone)
	if len(key) != 11 || key[0] != '1' {
		return fmt.Errorf("请输入正确的11位大陆手机号")
	}
	if err := SMSCanSend(phone); err != nil {
		return err
	}
	code := genSMSCode()
	exp := time.Duration(cfg.SMS.CodeExpireMinutes) * time.Minute
	if exp <= 0 {
		exp = 5 * time.Minute
	}

	cred := common.NewCredential(cfg.SMS.SecretID, cfg.SMS.SecretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"
	client, _ := sms.NewClient(cred, "ap-guangzhou", cpf)
	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = common.StringPtr(cfg.SMS.SmsSdkAppID)
	req.SignName = common.StringPtr(cfg.SMS.SignName)
	req.TemplateId = common.StringPtr(cfg.SMS.TemplateID)
	req.PhoneNumberSet = []*string{common.StringPtr("+86" + key)}
	req.TemplateParamSet = []*string{common.StringPtr(code), common.StringPtr(fmt.Sprintf("%d", cfg.SMS.CodeExpireMinutes))}

	_, err := client.SendSms(req)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("短信发送失败: %v", err)
	}
	if err != nil {
		return err
	}
	SMSSetCode(phone, code, exp)
	_ = context.Background()
	return nil
}

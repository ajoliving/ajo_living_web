/*
 * Alibaba Cloud domestic SMS OTP provider.
 * 1. Build the official Alibaba Cloud SMS client from runtime configuration.
 * 2. Send mainland China verification codes through an approved sign and template.
 * 3. Keep provider errors free of credentials, verification codes, and phone numbers.
 */
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v5/client"
	"github.com/alibabacloud-go/tea/dara"

	"ajoliving_web/http_service/internal/config"
)

// 1. aliyunSMSSender isolates the SDK request for provider tests.
type aliyunSMSSender interface {
	Send(ctx context.Context, phone string, signName string, templateCode string, templateParam string) error
}

// 2. AliyunSMSOTPProvider sends domestic verification codes through Alibaba Cloud SMS.
type AliyunSMSOTPProvider struct {
	sender            aliyunSMSSender
	signName          string
	templateCode      string
	templateParamName string
}

// 3. aliyunSMSClient adapts the generated Alibaba Cloud SDK client.
type aliyunSMSClient struct {
	client *dysmsapi.Client
}

// 4. newAliyunSMSOTPProvider builds a mainland SMS provider from runtime configuration.
func newAliyunSMSOTPProvider(cfg *config.Config) (OTPProvider, error) {
	if err := validateAliyunSMSConfig(cfg); err != nil {
		return nil, err
	}

	timeoutMilliseconds := int(cfg.AliyunSMSRequestTimeout.Milliseconds())
	client, err := dysmsapi.NewClient(&openapiutil.Config{
		AccessKeyId:     dara.String(cfg.AliyunSMSAccessKeyID),
		AccessKeySecret: dara.String(cfg.AliyunSMSAccessKeySecret),
		RegionId:        dara.String(cfg.AliyunSMSRegionID),
		ConnectTimeout:  dara.Int(timeoutMilliseconds),
		ReadTimeout:     dara.Int(timeoutMilliseconds),
	})
	if err != nil {
		return nil, fmt.Errorf("create Alibaba Cloud SMS client: %w", err)
	}

	return &AliyunSMSOTPProvider{
		sender:            &aliyunSMSClient{client: client},
		signName:          strings.TrimSpace(cfg.AliyunSMSSignName),
		templateCode:      strings.TrimSpace(cfg.AliyunSMSTemplateCode),
		templateParamName: strings.TrimSpace(cfg.AliyunSMSTemplateParamName),
	}, nil
}

// 5. SendCode sends one mainland China verification code.
func (p *AliyunSMSOTPProvider) SendCode(ctx context.Context, phone string, _ string, code string) error {
	normalizedPhone, err := normalizeMainlandSMSPhone(phone)
	if err != nil {
		return err
	}

	templateParam, err := json.Marshal(map[string]string{p.templateParamName: strings.TrimSpace(code)})
	if err != nil {
		return fmt.Errorf("marshal Alibaba Cloud SMS template parameter: %w", err)
	}

	return p.sender.Send(ctx, normalizedPhone, p.signName, p.templateCode, string(templateParam))
}

// 6. Send issues one Alibaba Cloud SendSms request and checks its service status.
func (c *aliyunSMSClient) Send(ctx context.Context, phone string, signName string, templateCode string, templateParam string) error {
	response, err := c.client.SendSmsWithContext(ctx, &dysmsapi.SendSmsRequest{
		PhoneNumbers:  dara.String(phone),
		SignName:      dara.String(signName),
		TemplateCode:  dara.String(templateCode),
		TemplateParam: dara.String(templateParam),
	}, &dara.RuntimeOptions{})
	if err != nil {
		return fmt.Errorf("Alibaba Cloud SMS request failed: %w", err)
	}
	if response == nil || response.Body == nil || !strings.EqualFold(strings.TrimSpace(dara.StringValue(response.Body.Code)), "OK") {
		return fmt.Errorf("Alibaba Cloud SMS rejected the request")
	}

	return nil
}

// 7. validateAliyunSMSConfig checks required domestic SMS configuration.
func validateAliyunSMSConfig(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("Alibaba Cloud SMS configuration is required")
	}
	if strings.TrimSpace(cfg.AliyunSMSAccessKeyID) == "" || strings.TrimSpace(cfg.AliyunSMSAccessKeySecret) == "" || strings.TrimSpace(cfg.AliyunSMSSignName) == "" || strings.TrimSpace(cfg.AliyunSMSTemplateCode) == "" {
		return fmt.Errorf("Alibaba Cloud SMS credentials, sign name, and mainland template code are required")
	}
	if strings.TrimSpace(cfg.AliyunSMSRegionID) == "" || strings.TrimSpace(cfg.AliyunSMSTemplateParamName) == "" || cfg.AliyunSMSRequestTimeout <= 0 {
		return fmt.Errorf("Alibaba Cloud SMS region, template parameter name, and timeout are required")
	}

	return nil
}

// 8. normalizeMainlandSMSPhone validates one E.164 mainland China mobile number.
func normalizeMainlandSMSPhone(phone string) (string, error) {
	normalized := strings.NewReplacer(" ", "", "-", "").Replace(strings.TrimSpace(phone))
	if !strings.HasPrefix(normalized, "+86") {
		return "", fmt.Errorf("Alibaba Cloud domestic SMS only supports +86 mobile numbers")
	}

	localNumber := strings.TrimPrefix(normalized, "+86")
	if len(localNumber) != 11 || localNumber[0] != '1' {
		return "", fmt.Errorf("invalid mainland China mobile number")
	}
	for _, value := range localNumber {
		if value < '0' || value > '9' {
			return "", fmt.Errorf("invalid mainland China mobile number")
		}
	}

	return "+86" + localNumber, nil
}

/*
 * Alibaba Cloud domestic SMS OTP tests.
 * 1. Verify approved signature, template, phone, and template parameters are forwarded.
 * 2. Verify mainland-only phone validation rejects unsupported destinations.
 * 3. Verify phone OTP delivery, cooldown, and one-time verification behavior.
 */
package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"ajoliving_web/http_service/internal/config"
)

// 1. recordedAliyunSMSRequest captures a provider request without calling Alibaba Cloud.
type recordedAliyunSMSRequest struct {
	phone         string
	signName      string
	templateCode  string
	templateParam string
}

// 2. fakeAliyunSMSSender records generated SDK input for tests.
type fakeAliyunSMSSender struct {
	requests []recordedAliyunSMSRequest
	err      error
}

// 3. Send records the request or returns the configured error.
func (s *fakeAliyunSMSSender) Send(_ context.Context, phone string, signName string, templateCode string, templateParam string) error {
	s.requests = append(s.requests, recordedAliyunSMSRequest{phone: phone, signName: signName, templateCode: templateCode, templateParam: templateParam})
	return s.err
}

// 4. fakeOTPProvider records generic OTP delivery requests.
type fakeOTPProvider struct {
	phones []string
	codes  []string
	err    error
}

// 5. SendCode records one OTP delivery request.
func (p *fakeOTPProvider) SendCode(_ context.Context, phone string, _ string, code string) error {
	p.phones = append(p.phones, phone)
	p.codes = append(p.codes, code)
	return p.err
}

// 6. TestAliyunSMSOTPProviderSendsApprovedTemplate verifies domestic provider request fields.
func TestAliyunSMSOTPProviderSendsApprovedTemplate(t *testing.T) {
	sender := &fakeAliyunSMSSender{}
	provider := &AliyunSMSOTPProvider{sender: sender, signName: "AJO Test", templateCode: "SMS_509205138", templateParamName: "code"}
	if err := provider.SendCode(context.Background(), "+86 138-1234-5678", "login", "123456"); err != nil {
		t.Fatalf("send code: %v", err)
	}
	if len(sender.requests) != 1 {
		t.Fatalf("expected one SMS request, got %d", len(sender.requests))
	}
	request := sender.requests[0]
	if request.phone != "+8613812345678" || request.signName != "AJO Test" || request.templateCode != "SMS_509205138" {
		t.Fatalf("unexpected Alibaba Cloud SMS request: %#v", request)
	}
	params := map[string]string{}
	if err := json.Unmarshal([]byte(request.templateParam), &params); err != nil {
		t.Fatalf("decode template parameter: %v", err)
	}
	if params["code"] != "123456" {
		t.Fatalf("unexpected template parameter: %#v", params)
	}
}

// 7. TestAliyunSMSOTPProviderRejectsUnsupportedCountry verifies domestic-only delivery.
func TestAliyunSMSOTPProviderRejectsUnsupportedCountry(t *testing.T) {
	sender := &fakeAliyunSMSSender{}
	provider := &AliyunSMSOTPProvider{sender: sender, signName: "test", templateCode: "SMS_test", templateParamName: "code"}
	if err := provider.SendCode(context.Background(), "+85291234567", "login", "123456"); err == nil {
		t.Fatal("expected Hong Kong phone to be rejected by domestic SMS provider")
	}
	if len(sender.requests) != 0 {
		t.Fatal("unsupported phone number must not reach Alibaba Cloud")
	}
}

// 8. TestAuthServicePhoneOTPDeliveryAndCooldown verifies secure OTP persistence and one-time use.
func TestAuthServicePhoneOTPDeliveryAndCooldown(t *testing.T) {
	now := time.Date(2026, time.July, 15, 10, 0, 0, 0, time.UTC)
	runtimeValue := newAuthTestRuntime(t, &config.Config{OTPProvider: "aliyun_sms", OTPResendCooldown: time.Minute})
	runtimeValue.Now = func() time.Time { return now }
	provider := &fakeOTPProvider{}
	runtimeValue.OTPProvider = provider
	authService := NewAuthService(runtimeValue)

	requested, err := authService.RequestOTP(context.Background(), RequestOTPParams{PhoneCountryCode: "+86", PhoneNumber: "13812345678", Scene: "login"})
	if err != nil {
		t.Fatalf("request OTP: %v", err)
	}
	if requested.MockCode != "" || requested.ExpiresIn != 300 || len(provider.codes) != 1 || len(provider.codes[0]) != 6 {
		t.Fatalf("unexpected OTP delivery result: %#v %#v", requested, provider)
	}
	for _, value := range provider.codes[0] {
		if value < '0' || value > '9' {
			t.Fatalf("OTP code must be numeric, got %q", provider.codes[0])
		}
	}
	if _, err := authService.RequestOTP(context.Background(), RequestOTPParams{PhoneCountryCode: "+86", PhoneNumber: "13812345678", Scene: "login"}); !hasAppErrorCode(err, "RATE_LIMITED") {
		t.Fatalf("expected phone cooldown, got %v", err)
	}

	verified, err := authService.VerifyOTP(context.Background(), VerifyOTPParams{PhoneCountryCode: "+86", PhoneNumber: "13812345678", Scene: "login", Code: provider.codes[0]})
	if err != nil || verified.AccessToken == "" {
		t.Fatalf("verify OTP: %#v %v", verified, err)
	}
	if _, err := authService.VerifyOTP(context.Background(), VerifyOTPParams{PhoneCountryCode: "+86", PhoneNumber: "13812345678", Scene: "login", Code: provider.codes[0]}); !hasAppErrorCode(err, "VALIDATION_ERROR") {
		t.Fatalf("expected consumed OTP to be rejected, got %v", err)
	}
}

// 9. TestAuthServiceDoesNotPersistFailedPhoneDelivery verifies failed delivery leaves no usable code.
func TestAuthServiceDoesNotPersistFailedPhoneDelivery(t *testing.T) {
	now := time.Date(2026, time.July, 15, 10, 0, 0, 0, time.UTC)
	runtimeValue := newAuthTestRuntime(t, &config.Config{OTPProvider: "aliyun_sms", OTPResendCooldown: time.Minute})
	runtimeValue.Now = func() time.Time { return now }
	provider := &fakeOTPProvider{err: errors.New("upstream unavailable")}
	runtimeValue.OTPProvider = provider
	authService := NewAuthService(runtimeValue)

	if _, err := authService.RequestOTP(context.Background(), RequestOTPParams{PhoneCountryCode: "+86", PhoneNumber: "13812345678"}); !hasAppErrorCode(err, "INTERNAL_ERROR") {
		t.Fatalf("expected delivery error, got %v", err)
	}
	if _, err := authService.VerifyOTP(context.Background(), VerifyOTPParams{PhoneCountryCode: "+86", PhoneNumber: "13812345678", Code: "123456"}); !hasAppErrorCode(err, "VALIDATION_ERROR") {
		t.Fatalf("failed delivery must not create a usable OTP, got %v", err)
	}
	provider.err = nil
	if _, err := authService.RequestOTP(context.Background(), RequestOTPParams{PhoneCountryCode: "+86", PhoneNumber: "13812345678"}); err != nil {
		t.Fatalf("failed delivery must release phone cooldown: %v", err)
	}
}

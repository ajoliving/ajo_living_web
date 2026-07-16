/*
 * 代理資料審核結果郵件。
 * 1. 讀取代理帳戶的註冊電郵並建立正式審核結果內容。
 * 2. 郵件發送失敗只記錄日誌，不回滾已完成的人工審核。
 */
package service

import (
	"context"
	"fmt"
	"strings"
)

// 1. sendAgencyProfileReviewEmail sends the committed review result to the registered email.
func (s *AgencyProfileService) sendAgencyProfileReviewEmail(ctx context.Context, profilePublicID string, approved bool, reviewNote string) {
	if s.runtime.MailSender == nil {
		return
	}

	var recipient struct {
		Email  *string
		NameZH string
		NameEN string
	}
	err := s.runtime.DB.WithContext(ctx).
		Table("agency_profiles").
		Select("user_credentials.email, agency_profiles.name_zh, agency_profiles.name_en").
		Joins("JOIN user_credentials ON user_credentials.user_id = agency_profiles.user_id").
		Where("agency_profiles.public_id = ?", strings.TrimSpace(profilePublicID)).
		Take(&recipient).Error
	if err != nil || recipient.Email == nil || strings.TrimSpace(*recipient.Email) == "" {
		return
	}

	subject, body := buildAgencyProfileReviewEmailContent(
		firstNonBlank(recipient.NameZH, recipient.NameEN),
		approved,
		strings.TrimSpace(reviewNote),
		strings.TrimSpace(s.runtime.Config.AppPublicBaseURL),
	)
	if err := s.runtime.MailSender.Send(ctx, strings.TrimSpace(*recipient.Email), subject, body); err != nil && s.runtime.Logger != nil {
		s.runtime.Logger.Warn("failed to send agency profile review email", "profile_public_id", profilePublicID, "error", err)
	}
}

// 2. buildAgencyProfileReviewEmailContent builds the Traditional Chinese review result email.
func buildAgencyProfileReviewEmailContent(name string, approved bool, reviewNote string, publicBaseURL string) (string, string) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "會員"
	}

	subject := "[AJO Living] 代理資料審核已通過"
	body := fmt.Sprintf("%s：\n\n您的代理資料已通過審核，現可登入 AJO Living 並使用已批准的代理資料刊登樓盤。", name)
	if !approved {
		subject = "[AJO Living] 代理資料審核未通過"
		body = fmt.Sprintf("%s：\n\n您的代理資料未通過審核，請登入 AJO Living 查看並修改資料後重新提交。\n\n拒絕原因：%s", name, firstNonBlank(reviewNote, "請登入會員中心查看審核意見。"))
	}

	if baseURL := strings.TrimRight(publicBaseURL, "/"); baseURL != "" {
		body += "\n\n查看代理資料：" + baseURL + "/account/profile/agency-profile"
	}
	body += "\n\nAJO Living"
	return subject, body
}

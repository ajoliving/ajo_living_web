/*
 * Staff wallet HTTP handlers.
 * 1. Bind staff-only wallet ledger, point grant, and reward ad requests.
 * 2. Delegate wallet business rules to the service layer.
 * 3. Keep staff responses aligned with the unified API envelope.
 */
package handler

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. StaffWalletHandler handles staff-only wallet endpoints.
type StaffWalletHandler struct {
	walletService *service.WalletService
}

// 2. staffWalletGrantRequest defines an operator point grant payload.
type staffWalletGrantRequest struct {
	UserID string `json:"user_id"`
	Amount int64  `json:"amount"`
	Note   string `json:"note"`
}

// 3. staffRewardAdCreateRequest defines a rewarded ad creation payload.
type staffRewardAdCreateRequest struct {
	Title         string `json:"title"`
	Summary       string `json:"summary"`
	CoverURL      string `json:"cover_url"`
	MediaURL      string `json:"media_url"`
	MediaType     string `json:"media_type"`
	TargetURL     string `json:"target_url"`
	RewardPoints  int64  `json:"reward_points"`
	WatchSeconds  int    `json:"watch_seconds"`
	TotalBudget   int64  `json:"total_budget"`
	IsActive      *bool  `json:"is_active"`
	StartsAt      string `json:"starts_at"`
	EndsAt        string `json:"ends_at"`
	RetentionDays int    `json:"retention_days"`
}

// 4. staffRewardAdUpdateRequest defines a rewarded ad update payload.
type staffRewardAdUpdateRequest struct {
	Title         *string `json:"title"`
	Summary       *string `json:"summary"`
	CoverURL      *string `json:"cover_url"`
	MediaURL      *string `json:"media_url"`
	MediaType     *string `json:"media_type"`
	TargetURL     *string `json:"target_url"`
	RewardPoints  *int64  `json:"reward_points"`
	WatchSeconds  *int    `json:"watch_seconds"`
	TotalBudget   *int64  `json:"total_budget"`
	IsActive      *bool   `json:"is_active"`
	StartsAt      *string `json:"starts_at"`
	EndsAt        *string `json:"ends_at"`
	RetentionDays *int    `json:"retention_days"`
}

// 5. NewStaffWalletHandler creates a staff wallet handler instance.
func NewStaffWalletHandler(walletService *service.WalletService) *StaffWalletHandler {
	return &StaffWalletHandler{walletService: walletService}
}

// 6. ListTransactions returns staff-visible wallet transactions.
func (h *StaffWalletHandler) ListTransactions(c *gin.Context) {
	page, pageSize := parsePagination(c)
	result, pagination, err := h.walletService.ListStaffWalletTransactions(c.Request.Context(), service.StaffWalletTransactionFilters{
		Page:       page,
		PageSize:   pageSize,
		UserID:     strings.TrimSpace(c.Query("user_id")),
		Direction:  strings.TrimSpace(c.Query("direction")),
		SourceType: strings.TrimSpace(c.Query("source_type")),
		BizModule:  strings.TrimSpace(c.Query("biz_module")),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": result, "pagination": pagination})
}

// 7. GrantPoints credits AJO Points to a selected member.
func (h *StaffWalletHandler) GrantPoints(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request staffWalletGrantRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.walletService.GrantOperatorPointsByUserID(
		c.Request.Context(),
		user.UserID,
		strings.TrimSpace(request.UserID),
		request.Amount,
		strings.TrimSpace(request.Note),
	)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 8. ListRewardAds returns staff-visible rewarded ad tasks.
func (h *StaffWalletHandler) ListRewardAds(c *gin.Context) {
	page, pageSize := parsePagination(c)
	result, pagination, err := h.walletService.ListStaffRewardAds(c.Request.Context(), service.StaffRewardAdFilters{
		Page:     page,
		PageSize: pageSize,
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		IsActive: parseOptionalBoolQuery(c, "is_active"),
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": result, "pagination": pagination})
}

// 9. GetRewardAd returns one staff-visible rewarded ad task.
func (h *StaffWalletHandler) GetRewardAd(c *gin.Context) {
	result, err := h.walletService.GetStaffRewardAd(c.Request.Context(), strings.TrimSpace(c.Param("taskId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 10. CreateRewardAd creates a rewarded ad task.
func (h *StaffWalletHandler) CreateRewardAd(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request staffRewardAdCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}
	startsAt, endsAt, err := parseRewardAdTimeRange(request.StartsAt, request.EndsAt)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	isActive := true
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	result, err := h.walletService.CreateStaffRewardAd(c.Request.Context(), user.UserID, service.RewardAdCreateParams{
		Title:         strings.TrimSpace(request.Title),
		Summary:       strings.TrimSpace(request.Summary),
		CoverURL:      strings.TrimSpace(request.CoverURL),
		MediaURL:      strings.TrimSpace(request.MediaURL),
		MediaType:     strings.TrimSpace(request.MediaType),
		TargetURL:     strings.TrimSpace(request.TargetURL),
		RewardPoints:  request.RewardPoints,
		WatchSeconds:  request.WatchSeconds,
		TotalBudget:   0,
		IsActive:      isActive,
		StartsAt:      startsAt,
		EndsAt:        endsAt,
		RetentionDays: request.RetentionDays,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 11. UpdateRewardAd updates a rewarded ad task.
func (h *StaffWalletHandler) UpdateRewardAd(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request staffRewardAdUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}
	params, err := rewardAdUpdateParams(request)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	result, err := h.walletService.UpdateStaffRewardAd(
		c.Request.Context(),
		user.UserID,
		strings.TrimSpace(c.Param("taskId")),
		params,
	)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}
	errcode.Success(c, result)
}

// 12. rewardAdUpdateParams converts a handler request to service params.
func rewardAdUpdateParams(request staffRewardAdUpdateRequest) (service.RewardAdUpdateParams, error) {
	params := service.RewardAdUpdateParams{
		Title:         trimmedStringPointer(request.Title),
		Summary:       trimmedStringPointer(request.Summary),
		CoverURL:      trimmedStringPointer(request.CoverURL),
		MediaURL:      trimmedStringPointer(request.MediaURL),
		MediaType:     trimmedStringPointer(request.MediaType),
		TargetURL:     trimmedStringPointer(request.TargetURL),
		RewardPoints:  request.RewardPoints,
		WatchSeconds:  request.WatchSeconds,
		TotalBudget:   request.TotalBudget,
		IsActive:      request.IsActive,
		RetentionDays: request.RetentionDays,
	}
	if request.StartsAt != nil {
		startsAt, err := parseOptionalRFC3339(*request.StartsAt)
		if err != nil {
			return params, err
		}
		params.StartsAtSet = true
		params.StartsAt = startsAt
	}
	if request.EndsAt != nil {
		endsAt, err := parseOptionalRFC3339(*request.EndsAt)
		if err != nil {
			return params, err
		}
		params.EndsAtSet = true
		params.EndsAt = endsAt
	}
	return params, nil
}

// 13. parseRewardAdTimeRange parses optional create time fields.
func parseRewardAdTimeRange(startsAt string, endsAt string) (*time.Time, *time.Time, error) {
	startTime, err := parseOptionalRFC3339(startsAt)
	if err != nil {
		return nil, nil, err
	}
	endTime, err := parseOptionalRFC3339(endsAt)
	if err != nil {
		return nil, nil, err
	}
	return startTime, endTime, nil
}

// 14. parseOptionalRFC3339 parses an RFC3339 string or returns nil for blank.
func parseOptionalRFC3339(value string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return nil, errcode.New(errcode.CodeValidationError, "time must use RFC3339 format")
	}
	return &parsed, nil
}

// 15. trimmedStringPointer trims optional string pointer fields.
func trimmedStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

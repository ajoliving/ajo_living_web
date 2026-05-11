/*
 * 首頁內容 HTTP 處理器。
 * 1. 綁定首頁輪播、三個主模組與登入背景圖設定請求。
 * 2. 輸出公開首頁內容與登入頁背景圖。
 * 3. 將保存邏輯交由 service 層處理。
 */
package handler

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. HomeContentHandler handles homepage content endpoints.
type HomeContentHandler struct {
	homeContentService *service.HomeContentService
}

// 2. saveHomeCarouselRequest defines carousel save payload.
type saveHomeCarouselRequest struct {
	Items []service.HomeCarouselInput `json:"items"`
}

// 3. saveHomeModuleCardsRequest defines module cards save payload.
type saveHomeModuleCardsRequest struct {
	Cards []service.HomeModuleCardInput `json:"cards"`
}

// 4. saveLoginHeroRequest defines login hero save payload.
type saveLoginHeroRequest struct {
	Items []service.LoginHeroInput `json:"items"`
}

// 5. NewHomeContentHandler creates a homepage content handler.
func NewHomeContentHandler(homeContentService *service.HomeContentService) *HomeContentHandler {
	return &HomeContentHandler{homeContentService: homeContentService}
}

// 6. PublicContent returns public homepage content.
func (h *HomeContentHandler) PublicContent(c *gin.Context) {
	result, err := h.homeContentService.GetPublicHomeContent(c.Request.Context())
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 7. SettingsCarousel returns saved carousel settings.
func (h *HomeContentHandler) SettingsCarousel(c *gin.Context) {
	result, err := h.homeContentService.ListCarouselSettings(c.Request.Context())
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": result})
}

// 8. SaveSettingsCarousel saves carousel settings.
func (h *HomeContentHandler) SaveSettingsCarousel(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request saveHomeCarouselRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.homeContentService.SaveCarouselSettings(c.Request.Context(), user.UserID, request.Items)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": result})
}

// 9. SettingsModuleCards returns saved module card settings.
func (h *HomeContentHandler) SettingsModuleCards(c *gin.Context) {
	result, err := h.homeContentService.ListModuleCardSettings(c.Request.Context())
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"cards": result})
}

// 10. SaveSettingsModuleCards saves module card settings.
func (h *HomeContentHandler) SaveSettingsModuleCards(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request saveHomeModuleCardsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.homeContentService.SaveModuleCardSettings(c.Request.Context(), user.UserID, request.Cards)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"cards": result})
}

// 11. LoginHero returns saved login page hero image.
func (h *HomeContentHandler) LoginHero(c *gin.Context) {
	result, err := h.homeContentService.ListLoginHeroSettings(c.Request.Context())
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": result})
}

// 12. SettingsLoginHero returns saved login page hero setting.
func (h *HomeContentHandler) SettingsLoginHero(c *gin.Context) {
	result, err := h.homeContentService.ListLoginHeroSettings(c.Request.Context())
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": result})
}

// 13. SaveSettingsLoginHero saves login page hero setting.
func (h *HomeContentHandler) SaveSettingsLoginHero(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request saveLoginHeroRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid request payload"))
		return
	}

	result, err := h.homeContentService.SaveLoginHeroSettings(c.Request.Context(), user.UserID, request.Items)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": result})
}

/*
 * Property content translation HTTP handler.
 * 1. Bind authenticated property translation requests.
 * 2. Delegate Traditional Chinese to English translation to the property service.
 * 3. Return translated title and description through the unified response format.
 */
package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. propertyContentTranslationRequest defines property translation input.
type propertyContentTranslationRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// 2. TranslatePropertyContent translates property title and description to English.
func (h *PropertyHandler) TranslatePropertyContent(c *gin.Context) {
	if currentUser(c) == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var request propertyContentTranslationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid property translation request"))
		return
	}

	result, err := h.propertyService.TranslatePropertyContent(
		c.Request.Context(),
		service.PropertyContentTranslationInput{
			Title:       strings.TrimSpace(request.Title),
			Description: strings.TrimSpace(request.Description),
		},
	)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

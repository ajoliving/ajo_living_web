/*
 * POS 樓宇只讀 HTTP 介面。
 * 1. 輸出註冊頁可用大廈清單。
 * 2. 輸出指定大廈下的單位清單。
 */
package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. POSBuildingHandler handles POS building metadata endpoints.
type POSBuildingHandler struct {
	posBuildingService *service.POSBuildingService
	posPaymentService  *service.POSPaymentService
}

// 2. NewPOSBuildingHandler creates a POS building handler instance.
func NewPOSBuildingHandler(posBuildingService *service.POSBuildingService, posPaymentService *service.POSPaymentService) *POSBuildingHandler {
	return &POSBuildingHandler{posBuildingService: posBuildingService, posPaymentService: posPaymentService}
}

// 3. ListBuildings returns POS building options.
func (h *POSBuildingHandler) ListBuildings(c *gin.Context) {
	result, err := h.posBuildingService.ListBuildings(c.Request.Context())
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": result})
}

// 4. ListUnits returns POS flat units for a building.
func (h *POSBuildingHandler) ListUnits(c *gin.Context) {
	result, err := h.posBuildingService.ListUnits(c.Request.Context(), strings.TrimSpace(c.Param("buildingId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": result})
}

// 5. ListMemberBuildings returns POS buildings through the current member token.
func (h *POSBuildingHandler) ListMemberBuildings(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.posPaymentService.ListMemberBuildings(c.Request.Context(), user.UserID)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": result})
}

// 6. ListMemberUnits returns POS units through the current member token.
func (h *POSBuildingHandler) ListMemberUnits(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	result, err := h.posPaymentService.ListMemberUnits(c.Request.Context(), user.UserID, strings.TrimSpace(c.Param("buildingId")))
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": result})
}

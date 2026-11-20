/*
 * 首頁社區環境讀取接口。
 * 1. 訪客使用全港資料，登入會員可帶目前大廈 ID。
 * 2. 位置與可見範圍由服務端解析，個人化回應不得由瀏覽器共用快取。
 */
package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. HomeEnvironmentHandler serves the homepage read-only environment summary.
type HomeEnvironmentHandler struct {
	Service *service.HomeEnvironmentService
}

// 2. Read validates public options and passes member identity only for active accounts.
func (h HomeEnvironmentHandler) Read(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	lang := c.DefaultQuery("lang", "tc")
	buildingID := strings.TrimSpace(c.Query("building_id"))
	estate := c.Query("estate")
	if (lang != "tc" && lang != "en") || len(buildingID) > 100 ||
		(estate != "" && (!service.ValidHomeEstate(estate) || buildingID != "")) {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid environment query"))
		return
	}
	var userID int64
	if user := currentUser(c); user != nil && user.MemberStatus == "active" {
		userID = user.UserID
	}
	errcode.Success(c, h.Service.Read(c.Request.Context(), userID, buildingID, lang, estate, userID > 0 && c.Query("detail") == "forecast"))
}

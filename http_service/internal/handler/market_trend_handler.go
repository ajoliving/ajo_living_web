/*
 * 市場走勢 HTTP 處理器。
 * 1. 輸出公開市場租金走勢資料。
 * 2. 將公開資料源讀取與解析交由 service 層處理。
 */
package handler

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. MarketTrendHandler handles public market trend endpoints.
type MarketTrendHandler struct {
	marketTrendService *service.MarketTrendService
}

// 2. NewMarketTrendHandler creates a market trend handler.
func NewMarketTrendHandler(marketTrendService *service.MarketTrendService) *MarketTrendHandler {
	return &MarketTrendHandler{marketTrendService: marketTrendService}
}

// 3. RentTrend returns public Hong Kong private domestic rent trend data.
func (h *MarketTrendHandler) RentTrend(c *gin.Context) {
	result, err := h.marketTrendService.GetRentTrend(c.Request.Context())
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

/*
 * Supermarket offer HTTP handlers.
 * 1. Expose public good-price summary, search, and product detail endpoints.
 * 2. Expose AJO member favorites and price alert endpoints.
 * 3. Keep old good-price account APIs hidden from AJO frontend.
 */
package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

// 1. SupermarketOfferHandler handles supermarket offer endpoints.
type SupermarketOfferHandler struct {
	supermarketService *service.SupermarketOfferService
}

// 2. NewSupermarketOfferHandler creates a supermarket offer handler.
func NewSupermarketOfferHandler(supermarketService *service.SupermarketOfferService) *SupermarketOfferHandler {
	return &SupermarketOfferHandler{supermarketService: supermarketService}
}

// 3. Summary returns public supermarket offer summary.
func (h *SupermarketOfferHandler) Summary(c *gin.Context) {
	result, err := h.supermarketService.Summary(c.Request.Context())
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 4. Search returns public supermarket product search results.
func (h *SupermarketOfferHandler) Search(c *gin.Context) {
	page, pageSize := supermarketPagination(c)
	result, err := h.supermarketService.Search(c.Request.Context(), service.SupermarketSearchFilters{
		Query:     c.Query("q"),
		Category:  c.Query("category"),
		Brand:     c.Query("brand"),
		Store:     c.Query("store"),
		OfferOnly: parseBoolQuery(c, "offerOnly") || parseBoolQuery(c, "offer_only"),
		Sort:      c.Query("sort"),
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 5. ProductDetail returns public product detail with optional member state.
func (h *SupermarketOfferHandler) ProductDetail(c *gin.Context) {
	var userID *int64
	if user := currentUser(c); user != nil {
		userID = &user.UserID
	}
	days := parsePositiveIntQuery(c, "days")
	result, err := h.supermarketService.ProductDetail(c.Request.Context(), c.Param("code"), days, userID)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 6. ListFavorites returns current member supermarket favorites.
func (h *SupermarketOfferHandler) ListFavorites(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	page, pageSize := supermarketPagination(c)
	items, pagination, err := h.supermarketService.ListFavorites(c.Request.Context(), user.UserID, page, pageSize)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": items, "pagination": pagination})
}

// 7. AddFavorite saves one current member supermarket favorite.
func (h *SupermarketOfferHandler) AddFavorite(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var req struct {
		ProductCode string `json:"productCode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid favorite payload"))
		return
	}
	result, err := h.supermarketService.AddFavorite(c.Request.Context(), user.UserID, req.ProductCode)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 8. RemoveFavorite deletes one current member supermarket favorite.
func (h *SupermarketOfferHandler) RemoveFavorite(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	if err := h.supermarketService.RemoveFavorite(c.Request.Context(), user.UserID, c.Param("code")); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"productCode": strings.TrimSpace(c.Param("code")), "isFavorite": false})
}

// 9. ListPriceAlerts returns current member supermarket price alerts.
func (h *SupermarketOfferHandler) ListPriceAlerts(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	items, err := h.supermarketService.ListPriceAlerts(c.Request.Context(), user.UserID)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"items": items})
}

// 10. CreatePriceAlert creates or replaces a member price alert.
func (h *SupermarketOfferHandler) CreatePriceAlert(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var req struct {
		ProductCode   string   `json:"productCode"`
		TargetPrice   *float64 `json:"targetPrice"`
		PriceMode     string   `json:"priceMode"`
		OfferRequired bool     `json:"offerRequired"`
		Enabled       *bool    `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid price alert payload"))
		return
	}
	result, err := h.supermarketService.CreatePriceAlert(c.Request.Context(), user.UserID, service.SupermarketPriceAlertInput{
		ProductCode:   req.ProductCode,
		TargetPrice:   req.TargetPrice,
		PriceMode:     req.PriceMode,
		OfferRequired: req.OfferRequired,
		Enabled:       req.Enabled,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 11. UpdatePriceAlert updates one member price alert.
func (h *SupermarketOfferHandler) UpdatePriceAlert(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	var req struct {
		TargetPrice   *float64 `json:"targetPrice"`
		PriceMode     string   `json:"priceMode"`
		OfferRequired *bool    `json:"offerRequired"`
		Enabled       *bool    `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid price alert payload"))
		return
	}
	result, err := h.supermarketService.UpdatePriceAlert(c.Request.Context(), user.UserID, supermarketAlertID(c), service.SupermarketPriceAlertUpdateInput{
		TargetPrice:   req.TargetPrice,
		PriceMode:     req.PriceMode,
		OfferRequired: req.OfferRequired,
		Enabled:       req.Enabled,
	})
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 12. DeletePriceAlert deletes one member price alert.
func (h *SupermarketOfferHandler) DeletePriceAlert(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
		return
	}

	if err := h.supermarketService.DeletePriceAlert(c.Request.Context(), user.UserID, supermarketAlertID(c)); err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, gin.H{"id": supermarketAlertID(c), "deleted": true})
}

// 13. CreateImageReport saves one public supermarket product image issue report.
func (h *SupermarketOfferHandler) CreateImageReport(c *gin.Context) {
	var req struct {
		ProductCode string `json:"productCode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		errcode.WriteError(c, errcode.New(errcode.CodeValidationError, "invalid image report payload"))
		return
	}

	result, err := h.supermarketService.CreateImageReport(c.Request.Context(), req.ProductCode)
	if err != nil {
		errcode.WriteError(c, err)
		return
	}

	errcode.Success(c, result)
}

// 14. supermarketPagination reads both AJO and good-price pagination query names.
func supermarketPagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("page", "1")))
	pageSizeRaw := c.Query("pageSize")
	if strings.TrimSpace(pageSizeRaw) == "" {
		pageSizeRaw = c.Query("page_size")
	}
	pageSize, _ := strconv.Atoi(strings.TrimSpace(pageSizeRaw))
	if pageSize <= 0 {
		pageSize = 20
	}
	return page, pageSize
}

// 15. supermarketAlertID reads the route alert id.
func supermarketAlertID(c *gin.Context) int64 {
	value, _ := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	return value
}

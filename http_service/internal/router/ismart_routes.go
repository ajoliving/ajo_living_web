/*
 * iSmart proxy route registration.
 * 1. Keep all iSmart external API proxy routes in one file.
 * 2. Require AJO authentication before proxying to old iSmart services.
 * 3. Expose documented public integration payment query routes.
 */
package router

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
)

// 1. registerPublicIsmartIntegrationRoutes registers public iSmart payment integration routes.
func registerPublicIsmartIntegrationRoutes(
	api *gin.RouterGroup,
	ismartHandler *handler.IsmartExternalHandler,
) {
	api.GET("/integration/payments/unpaid-invoices/", ismartHandler.ListPaymentUnpaidInvoices)
	api.GET("/integration/payments/transactions/by-unit/", ismartHandler.ListPaymentTransactionsByUnit)
	api.GET("/integration/payments/transactions/by-date/", ismartHandler.ListPaymentTransactionsByDate)
}

// 2. registerIsmartRoutes registers authenticated iSmart proxy routes.
func registerIsmartRoutes(
	api *gin.RouterGroup,
	ismartHandler *handler.IsmartExternalHandler,
	requireAuth gin.HandlerFunc,
) {
	api.GET("/me/ismart/buildings", requireAuth, ismartHandler.ListBuildings)
	api.POST("/me/ismart/account-registration", requireAuth, ismartHandler.RegisterAccount)
	api.GET("/me/ismart/building-info", requireAuth, ismartHandler.GetBuildingInfo)
	api.GET("/me/ismart/management-fees", requireAuth, ismartHandler.ListManagementFees)
	api.GET("/me/ismart/other-fees", requireAuth, ismartHandler.ListOtherFees)
	api.GET("/me/ismart/notices", requireAuth, ismartHandler.ListBuildingNotices)
	api.POST("/me/ismart/building-comments", requireAuth, ismartHandler.SubmitBuildingComment)
	api.POST("/me/ismart/owner-binding-requests", requireAuth, ismartHandler.SubmitOwnerBindingRequest)
	api.GET("/me/ismart/subaccounts", requireAuth, ismartHandler.ListSubaccounts)
	api.POST("/me/ismart/subaccounts/grant", requireAuth, ismartHandler.GrantSubaccount)
	api.POST("/me/ismart/subaccounts/revoke", requireAuth, ismartHandler.RevokeSubaccount)
	api.GET("/me/ismart/building-access", requireAuth, ismartHandler.GetBuildingAccess)
	api.POST("/me/ismart/building-access/open-door", requireAuth, ismartHandler.OpenDoor)
	api.POST("/me/ismart/building-access/qrcode", requireAuth, ismartHandler.GenerateQRCode)
	api.POST("/me/ismart/pos-payment-to-ismart", requireAuth, ismartHandler.SubmitPOSPayment)
}

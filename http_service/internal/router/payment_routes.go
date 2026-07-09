/*
 * Payment route registration.
 * 1. Register POS building metadata routes for registration and member binding.
 * 2. Register POS bill, fee, order, history, accounting, and public callback routes.
 */
package router

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
)

// 1. registerPublicPOSPaymentRoutes registers public POS metadata and callback routes.
func registerPublicPOSPaymentRoutes(
	api *gin.RouterGroup,
	posBuildingHandler *handler.POSBuildingHandler,
	walletHandler *handler.WalletHandler,
) {
	api.GET("/pos/buildings", posBuildingHandler.ListBuildings)
	api.GET("/pos/buildings/:buildingId/units", posBuildingHandler.ListUnits)
	api.POST("/payments/easylink/notify", walletHandler.HandlePaymentNotify)
}

// 2. registerPOSPaymentRoutes registers authenticated POS payment center routes.
func registerPOSPaymentRoutes(
	api *gin.RouterGroup,
	posBuildingHandler *handler.POSBuildingHandler,
	posPaymentHandler *handler.POSPaymentHandler,
	requireAuth gin.HandlerFunc,
) {
	api.GET("/me/pos/buildings", requireAuth, posBuildingHandler.ListMemberBuildings)
	api.GET("/me/pos/buildings/:buildingId/units", requireAuth, posBuildingHandler.ListMemberUnits)
	api.GET("/me/payments/pos/overview", requireAuth, posPaymentHandler.Overview)
	api.GET("/me/payments/pos/fees", requireAuth, posPaymentHandler.ListFees)
	api.GET("/me/payments/pos/bank-accounts", requireAuth, posPaymentHandler.ListBankAccounts)
	api.POST("/me/payments/pos/payments/report", requireAuth, posPaymentHandler.ReportPayment)
	api.POST("/me/payments/pos/terminal/pay", requireAuth, posPaymentHandler.PayByTerminal)
	api.GET("/me/payments/pos/bills", requireAuth, posPaymentHandler.ListBills)
	api.GET("/me/payments/pos/orders", requireAuth, posPaymentHandler.ListOrders)
	api.POST("/me/payments/pos/orders", requireAuth, posPaymentHandler.CreateOrder)
	api.POST("/me/payments/pos/orders/query", requireAuth, posPaymentHandler.QueryOrder)
	api.GET("/me/payments/pos/orders/:mchOrderNo", requireAuth, posPaymentHandler.GetOrder)
	api.POST("/me/payments/pos/orders/:mchOrderNo/close", requireAuth, posPaymentHandler.CloseOrder)
	api.POST("/me/payments/pos/orders/:mchOrderNo/cancel", requireAuth, posPaymentHandler.CancelOrder)
	api.POST("/me/payments/pos/orders/:mchOrderNo/simulate", requireAuth, posPaymentHandler.SimulateOrder)
	api.GET("/me/payments/pos/history", requireAuth, posPaymentHandler.ListHistory)
	api.GET("/me/payments/pos/history/:paymentId", requireAuth, posPaymentHandler.GetHistoryDetail)
	api.GET("/me/payments/pos/accounting", requireAuth, posPaymentHandler.ListAccounting)
	api.POST("/me/payments/pos/accounting/clear", requireAuth, posPaymentHandler.ClearAccounting)
	api.GET("/me/payments/pos/accounting/records", requireAuth, posPaymentHandler.ListAccountingRecords)
	api.GET("/me/payments/pos/accounting/records/:recordId", requireAuth, posPaymentHandler.GetAccountingRecord)
}

/*
 * Wallet route registration.
 * 1. Register member wallet, ad task, transaction, and recharge order routes.
 * 2. Keep wallet routes separate from POS property fee payment routes.
 */
package router

import (
	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/handler"
)

// 1. registerWalletRoutes registers member wallet routes.
func registerWalletRoutes(api *gin.RouterGroup, walletHandler *handler.WalletHandler, requireAuth gin.HandlerFunc) {
	api.GET("/me/wallet", requireAuth, walletHandler.GetWallet)
	api.GET("/me/wallet/transactions", requireAuth, walletHandler.ListTransactions)
	api.GET("/me/wallet/ad-tasks", requireAuth, walletHandler.ListAdTasks)
	api.POST("/me/wallet/ad-tasks/:taskId/start", requireAuth, walletHandler.StartAdTask)
	api.POST("/me/wallet/ad-tasks/:taskId/claim", requireAuth, walletHandler.ClaimAdTask)
	api.POST("/me/wallet/ad-tasks/:taskId/click", requireAuth, walletHandler.TrackAdTaskClick)
	api.POST("/me/wallet/recharge-orders", requireAuth, walletHandler.CreateRechargeOrder)
	api.GET("/me/wallet/recharge-orders/:orderId", requireAuth, walletHandler.GetRechargeOrder)
}

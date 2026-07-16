/*
 * Agency profile route registration.
 * 1. Register restricted-session profile routes.
 * 2. Register staff review routes.
 */
package router

import (
	"ajoliving_web/http_service/internal/handler"
	"github.com/gin-gonic/gin"
)

// 1. registerAgencyCompanyMemberRoutes registers unified member profile routes.
func registerAgencyCompanyMemberRoutes(api *gin.RouterGroup, h *handler.AgencyCompanyHandler, requireAuth gin.HandlerFunc) {
	api.GET("/me/agency-profile", requireAuth, h.GetMemberAgencyProfile)
	api.POST("/me/agency-profile", requireAuth, h.CreateMemberAgencyProfile)
	api.PATCH("/me/agency-profile", requireAuth, h.UpdateMemberAgencyProfile)
	api.POST("/me/agency-profile/submit", requireAuth, h.SubmitMemberAgencyProfile)
	api.GET("/me/agency-profile/subaccounts", requireAuth, h.ListSubaccounts)
	api.POST("/me/agency-profile/subaccounts", requireAuth, h.CreateSubaccount)
	api.PATCH("/me/agency-profile/subaccounts/:subaccountId/status", requireAuth, h.UpdateSubaccountStatus)
	api.DELETE("/me/agency-profile/subaccounts/:subaccountId", requireAuth, h.DeleteSubaccount)
}

// 2. registerAgencyCompanyStaffRoutes registers staff review routes.
func registerAgencyCompanyStaffRoutes(api *gin.RouterGroup, h *handler.AgencyCompanyHandler, requireStaff gin.HandlerFunc) {
	api.GET("/staff/agency-profiles", requireStaff, h.ListAgencyProfilesForStaff)
	api.GET("/staff/agency-profiles/:profileId", requireStaff, h.GetAgencyProfileForStaff)
	api.POST("/staff/agency-profiles/:profileId/review", requireStaff, h.ReviewAgencyProfile)
}

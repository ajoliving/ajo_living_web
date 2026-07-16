/*
 * Authentication middleware.
 * 1. Parse bearer tokens and load the current user identity.
 * 2. Support optional and required auth flows.
 * 3. Store auth identity in gin context for handlers.
 */
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ajoliving_web/http_service/internal/errcode"
	"ajoliving_web/http_service/internal/service"
)

const currentUserKey = "current_user"

// 1. CurrentUser stores the authenticated member identity.
type CurrentUser struct {
	UserID             int64
	PublicID           string
	MemberStatus       string
	MemberType         string
	IsStaff            bool
	Role               string
	PrimaryCommunityID *int64
	AccountType        string
}

// 2. OptionalAuth loads the user identity when a token is present.
func OptionalAuth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.Next()
			return
		}

		identity, err := authService.AuthenticateToken(c.Request.Context(), token)
		if err != nil {
			errcode.WriteError(c, err)
			c.Abort()
			return
		}

		setCurrentUser(c, identity)
		c.Next()
	}
}

// 3. RequireAuth blocks requests without a valid token.
func RequireAuth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
			c.Abort()
			return
		}

		identity, err := authService.AuthenticateToken(c.Request.Context(), token)
		if err != nil {
			errcode.WriteError(c, err)
			c.Abort()
			return
		}

		setCurrentUser(c, identity)
		c.Next()
	}
}

// 4. RequireActiveMember blocks restricted, rejected, and disabled member sessions.
func RequireActiveMember(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
			c.Abort()
			return
		}
		identity, err := authService.AuthenticateToken(c.Request.Context(), token)
		if err != nil {
			errcode.WriteError(c, err)
			c.Abort()
			return
		}
		if identity.MemberStatus != "active" {
			errcode.WriteError(c, errcode.New(errcode.CodeAuthForbidden, "agency profile approval is required"))
			c.Abort()
			return
		}
		setCurrentUser(c, identity)
		c.Next()
	}
}

// 5. RequireStaff blocks requests without a valid staff identity.
func RequireStaff(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			errcode.WriteError(c, errcode.New(errcode.CodeAuthRequired, "login required"))
			c.Abort()
			return
		}

		identity, err := authService.AuthenticateToken(c.Request.Context(), token)
		if err != nil {
			errcode.WriteError(c, err)
			c.Abort()
			return
		}

		if !identity.IsStaff {
			errcode.WriteError(c, errcode.New(errcode.CodeAuthForbidden, "staff access required"))
			c.Abort()
			return
		}

		setCurrentUser(c, identity)
		c.Next()
	}
}

// 6. GetCurrentUser returns the current user from gin context.
func GetCurrentUser(c *gin.Context) *CurrentUser {
	value, exists := c.Get(currentUserKey)
	if !exists {
		return nil
	}

	currentUser, _ := value.(*CurrentUser)
	return currentUser
}

// 7. setCurrentUser stores the authenticated identity on gin context.
func setCurrentUser(c *gin.Context, identity *service.AuthIdentity) {
	c.Set(currentUserKey, &CurrentUser{
		UserID:             identity.UserID,
		PublicID:           identity.PublicID,
		MemberStatus:       identity.MemberStatus,
		MemberType:         identity.MemberType,
		IsStaff:            identity.IsStaff,
		Role:               identity.Role,
		PrimaryCommunityID: identity.PrimaryCommunityID,
		AccountType:        identity.AccountType,
	})
}

// 8. extractBearerToken extracts the raw token from the Authorization header.
func extractBearerToken(header string) string {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

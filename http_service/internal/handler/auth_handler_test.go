/*
 * Authentication handler contract tests.
 * 1. Validate unified login request binding.
 * 2. Keep validation responses on the shared error envelope.
 */
package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// 1. TestAuthHandlerLoginIdentifierRejectsMissingIdentifier validates the request contract.
func TestAuthHandlerLoginIdentifierRejectsMissingIdentifier(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"password":"password123"}`))
	context.Request.Header.Set("Content-Type", "application/json")

	NewAuthHandler(nil).LoginIdentifier(context)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"code":"VALIDATION_ERROR"`) {
		t.Fatalf("unexpected validation response: %d %s", recorder.Code, recorder.Body.String())
	}
}

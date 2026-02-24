package handler

import (
	"net/http"

	"github.com/cachatto/master-slave-server/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OTCHandler handles One-Time-Code handshake HTTP endpoints.
type OTCHandler struct {
	otcService *service.OTCService
}

// NewOTCHandler creates a new OTCHandler.
func NewOTCHandler(otcService *service.OTCService) *OTCHandler {
	return &OTCHandler{otcService: otcService}
}

// ExchangeCodeRequest is the expected JSON body for POST /auth/exchange-code.
type ExchangeCodeRequest struct {
	AppID string `json:"app_id" binding:"required"`
}

// ClaimTokenRequest is the expected JSON body for POST /auth/claim-token.
type ClaimTokenRequest struct {
	Code      string `json:"code" binding:"required"`
	PackageID string `json:"package_id" binding:"required"`
}

// ExchangeCode handles POST /auth/exchange-code
// Requires a valid access token (via JWT middleware).
// Generates a short-lived one-time code for a specific slave app.
func (h *OTCHandler) ExchangeCode(c *gin.Context) {
	var req ExchangeCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "app_id is required",
		})
		return
	}

	// Parse the app ID
	appID, err := uuid.Parse(req.AppID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid app_id format",
		})
		return
	}

	// Get the authenticated user's ID from the JWT middleware context
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}
	userID := userIDVal.(uuid.UUID)

	result, err := h.otcService.ExchangeCode(userID, appID)
	if err != nil {
		status := http.StatusInternalServerError
		switch err {
		case service.ErrAppNotFound:
			status = http.StatusNotFound
		case service.ErrNoPermission:
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":       result.Code,
		"expires_at": result.ExpiresAt,
	})
}

// ClaimToken handles POST /auth/claim-token
// Does NOT require authentication (the code IS the authentication).
// Validates the one-time code and returns a JWT token pair.
func (h *OTCHandler) ClaimToken(c *gin.Context) {
	var req ClaimTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "code and package_id are required",
		})
		return
	}

	tokens, err := h.otcService.ClaimToken(req.Code, req.PackageID)
	if err != nil {
		status := http.StatusUnauthorized
		switch err {
		case service.ErrAppNotFound:
			status = http.StatusNotFound
		case service.ErrAppMismatch:
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "token claimed successfully",
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})
}

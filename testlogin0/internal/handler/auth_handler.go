package handler

import (
	"net/http"

	"login/internal/service"
	"login/pkg/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) GetClientID(c *gin.Context) {
	clientID, err := h.authService.GetClientID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get client ID"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"client_id": clientID})
}

func (h *AuthHandler) VerifyGoogleToken(c *gin.Context) {
	var req struct {
		IDToken string `json:"id_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authResponse, err := h.authService.VerifyGoogleToken(c.Request.Context(), req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.SetCookie("token", authResponse.AccessToken, 3600, "/", "localhost", false, true)

	c.JSON(http.StatusOK, authResponse)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if err := h.authService.Logout(c.Request.Context(), userID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}
	c.Status(http.StatusNoContent)
}

// func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
// 	userID, _ := c.Get("user_id")
// 	user, err := h.authService.GetUserByID(c.Request.Context(), userID.(string))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
// 		return
// 	}
// 	c.JSON(http.StatusOK, user)
// }

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	token, err := c.Cookie("token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid token"})
		return
	}

	userID, err := utils.ParseToken(token, h.authService.Cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		return
	}

	c.JSON(http.StatusOK, user)
}

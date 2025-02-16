package handlers

import (
	"context"
	"github.com/gratefultolord/merch-store/internal/services"
	"github.com/labstack/echo/v4"
	"net/http"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type AuthRequest struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

func (h *AuthHandler) Auth(c echo.Context) error {
	var req AuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "username and password are required"})
	}

	token, err := h.authService.Auth(context.Background(), req.Username, req.Password)
	if err != nil {
		c.Logger().Errorf("auth service error: %v, username: %s", err, req.Username)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

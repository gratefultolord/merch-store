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

func (h *AuthHandler) Auth(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "username and password are required"})
	}

	token, err := h.authService.Auth(context.Background(), username, password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

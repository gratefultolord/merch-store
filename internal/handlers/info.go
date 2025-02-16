package handlers

import (
	"context"
	"github.com/gratefultolord/merch-store/internal/services"
	"github.com/labstack/echo/v4"
	"net/http"
)

type InfoHandler struct {
	infoService services.InfoService
}

func NewInfoHandler(infoService services.InfoService) *InfoHandler {
	return &InfoHandler{
		infoService: infoService,
	}
}

func (h *InfoHandler) Info(c echo.Context) error {
	userIDInterface := c.Get("userID")
	if userIDInterface == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "user ID not found"})
	}

	userID, ok := userIDInterface.(int)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid user ID type"})
	}

	info, err := h.infoService.GetUserInfo(context.Background(), userID)
	if err != nil {
		c.Logger().Errorf("info service error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to get user"})
	}
	if info == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "user not found"})
	}

	return c.JSON(http.StatusOK, info)
}

package handlers

import (
	"context"
	"github.com/gratefultolord/merch-store/internal/repository"
	"github.com/gratefultolord/merch-store/internal/services"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type SendCoinHandler struct {
	coinService services.CoinService
	userRepo    repository.UserRepo
}

func NewSendCoinHandler(
	coinService services.CoinService,
	userRepo repository.UserRepo,
) *SendCoinHandler {
	return &SendCoinHandler{
		coinService: coinService,
		userRepo:    userRepo,
	}
}

func (h *SendCoinHandler) SendCoin(c echo.Context) error {
	fromUserIDInterface := c.Get("userID")
	if fromUserIDInterface == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "user ID not found in context",
		})
	}

	fromUserID, ok := fromUserIDInterface.(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "user ID not found in context",
		})
	}

	toUsername := c.FormValue("toUser")
	amountStr := c.FormValue("amount")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid amount",
		})
	}

	toUser, err := h.userRepo.GetByUsername(context.Background(), toUsername)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to get user",
		})
	}
	if toUser == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "toUser not found",
		})
	}

	if err := h.coinService.Send(context.Background(), fromUserID, toUser.ID, amount); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to send coin",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "coins sent successfully",
	})
}

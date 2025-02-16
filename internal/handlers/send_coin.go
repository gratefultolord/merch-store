package handlers

import (
	"context"
	"fmt"
	"github.com/gratefultolord/merch-store/internal/repository"
	"github.com/gratefultolord/merch-store/internal/services"
	"github.com/labstack/echo/v4"
	"net/http"
)

type SendCoinRequest struct {
	ToUser string `json:"toUser" validate:"required"`
	Amount int    `json:"amount" validate:"required,gte=1"`
}

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
	fmt.Printf("fromUserID: %v\n", fromUserID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "user ID not found in context",
		})
	}

	var req SendCoinRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "failed to parse body",
		})
	}

	fmt.Printf("SendCoinRequest: %+v\n", req)

	toUser, err := h.userRepo.GetByUsername(context.Background(), req.ToUser)
	fmt.Printf("toUser: %v\n", toUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed getting receiver",
		})
	}

	if toUser == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "toUser not found",
		})
	}

	if err := h.coinService.Send(context.Background(), fromUserID, toUser.ID, req.Amount); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed sending coin",
		})
	}

	return c.NoContent(http.StatusOK)
}

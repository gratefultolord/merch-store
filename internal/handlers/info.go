package handlers

import (
	"context"
	"github.com/gratefultolord/merch-store/internal/repository"
	"github.com/labstack/echo/v4"
	"net/http"
)

type InfoHandler struct {
	userRepo        repository.UserRepo
	transactionRepo repository.TransactionRepo
}

func NewInfoHandler(userRepo repository.UserRepo, transactionRepo repository.TransactionRepo) *InfoHandler {
	return &InfoHandler{
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
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
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "invalid user ID type"})
	}

	user, err := h.userRepo.GetByID(context.Background(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to get user"})
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "user not found"})
	}

	transactions, err := h.transactionRepo.GetByUserID(context.Background(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to get transactions"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"coins":        user.Balance,
		"inventory":    user.Inventory,
		"transactions": transactions,
	})
}

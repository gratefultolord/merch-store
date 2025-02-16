package handlers

import (
	"context"
	"github.com/gratefultolord/merch-store/internal/repository"
	"github.com/gratefultolord/merch-store/internal/services"
	"github.com/labstack/echo/v4"
	"net/http"
)

type BuyHandler struct {
	inventoryService services.InventoryService
	userRepo         repository.UserRepo
	itemRepo         repository.ItemRepo
}

func NewBuyHandler(
	inventoryService services.InventoryService,
	userRepo repository.UserRepo,
	itemRepo repository.ItemRepo,
) *BuyHandler {
	return &BuyHandler{
		inventoryService: inventoryService,
		userRepo:         userRepo,
		itemRepo:         itemRepo,
	}
}

func (h *BuyHandler) Buy(c echo.Context) error {
	userIDInterface := c.Get("userID")
	if userIDInterface == nil {
		return c.JSON(
			http.StatusUnauthorized,
			map[string]string{
				"error": "user ID not found in request context",
			})
	}

	userID, ok := userIDInterface.(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "invalid user ID type",
		})
	}

	itemName := c.Param("item")

	if itemName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "item name is required",
		})
	}

	if err := h.inventoryService.Buy(context.Background(), userID, itemName); err != nil {
		c.Logger().Errorf("buy service error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	return c.NoContent(http.StatusOK)
}

package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockInventoryService struct {
	mock.Mock
}

func (m *MockInventoryService) Buy(ctx context.Context, userID int, itemName string) error {
	args := m.Called(ctx, userID, itemName)
	return args.Error(0)
}

func TestBuyHandler_Buy(t *testing.T) {
	tests := []struct {
		name           string
		userID         interface{}
		itemName       string
		mockSetup      func(m *MockInventoryService)
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name:     "Successful buy",
			userID:   1,
			itemName: "sword",
			mockSetup: func(m *MockInventoryService) {
				m.On("Buy", mock.Anything, 1, "sword").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   nil,
		},
		{
			name:     "User not found",
			userID:   nil,
			itemName: "sword",
			mockSetup: func(m *MockInventoryService) {
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   map[string]string{"error": "user ID not found in request context"},
		},
		{
			name:     "Invalid user ID type",
			userID:   "not-an-int",
			itemName: "sword",
			mockSetup: func(m *MockInventoryService) {
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   map[string]string{"error": "invalid user ID type"},
		},
		{
			name:     "Item name is required",
			userID:   1,
			itemName: "",
			mockSetup: func(m *MockInventoryService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]string{"error": "item name is required"},
		},
		{
			name:     "Inventory service error",
			userID:   1,
			itemName: "sword",
			mockSetup: func(m *MockInventoryService) {
				m.On("Buy", mock.Anything, 1, "sword").Return(errors.New("services: insufficient balance"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]string{"error": "services: insufficient balance"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInventoryService := new(MockInventoryService)
			handler := NewBuyHandler(mockInventoryService, nil, nil)

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/buy/"+tt.itemName, nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.SetParamNames("item")
			c.SetParamValues(tt.itemName)

			if tt.userID != nil {
				c.Set("userID", tt.userID)
			}

			tt.mockSetup(mockInventoryService)

			err := handler.Buy(c)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != nil {
				var body map[string]string
				err := json.Unmarshal(rec.Body.Bytes(), &body)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, body)
			} else {
				assert.Empty(t, rec.Body.String())
			}

			mockInventoryService.AssertExpectations(t)
		})
	}
}

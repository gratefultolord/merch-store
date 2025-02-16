package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Auth(ctx context.Context, username, password string) (string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Error(1)
}

type errorBinder struct{}

func (e *errorBinder) Bind(i interface{}, c echo.Context) error {
	return errors.New("bind error")
}

func TestAuthHandler_BindError(t *testing.T) {
	mockAuthService := new(MockAuthService)
	handler := NewAuthHandler(mockAuthService)
	e := echo.New()
	e.Binder = &errorBinder{}
	req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewBufferString(`{}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Auth(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.JSONEq(t, `{"error": "invalid request body"}`, rec.Body.String())
}

func TestAuthHandler_Auth(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(m *MockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Successful authentication",
			requestBody: `{
				"username": "user1",
				"password": "password123"
			}`,
			mockSetup: func(m *MockAuthService) {
				m.On("Auth", mock.Anything, "user1", "password123").Return("expected_jwt_token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"token":"expected_jwt_token"}`,
		},
		{
			name: "Invalid JSON syntax",
			requestBody: `{
				"username": "user1",
			}`,
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request body"}`,
		},
		{
			name: "Empty credentials",
			requestBody: `{
				"username": "",
				"password": ""
			}`,
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"username and password are required"}`,
		},
		{
			name: "Auth service error",
			requestBody: `{
				"username": "user1",
				"password": "wrong_password"
			}`,
			mockSetup: func(m *MockAuthService) {
				m.On("Auth", mock.Anything, "user1", "wrong_password").Return("", errors.New("invalid credentials"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"invalid credentials"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthService := new(MockAuthService)
			tt.mockSetup(mockAuthService)

			handler := NewAuthHandler(mockAuthService)
			e := echo.New()

			req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewBufferString(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.Auth(c)

			assert.NoError(t, err, "Handler should not return error")
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			mockAuthService.AssertExpectations(t)
		})
	}
}

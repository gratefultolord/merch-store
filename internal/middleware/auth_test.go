package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/gratefultolord/merch-store/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gratefultolord/merch-store/mocks"
)

func TestNewAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		mockSetup      func(m *mocks.UserRepo)
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name:       "Successful authentication",
			authHeader: "VALID",
			mockSetup: func(m *mocks.UserRepo) {
				user := &models.User{
					ID:       1,
					Username: "user1",
					Balance:  1000,
				}
				m.On("GetByID", mock.Anything, 1).Return(user, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody:   nil,
		},
		{
			name:           "Missing authorization header",
			authHeader:     "",
			mockSetup:      func(m *mocks.UserRepo) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   map[string]string{"error": "missing authorization header"},
		},
		{
			name:           "Invalid authorization header format",
			authHeader:     "Bearer",
			mockSetup:      func(m *mocks.UserRepo) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   map[string]string{"error": "invalid authorization header format"},
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid_token",
			mockSetup:      func(m *mocks.UserRepo) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   map[string]string{"error": "invalid token"},
		},
		{
			name:           "Invalid token claims",
			authHeader:     "INVALID_CLAIMS",
			mockSetup:      func(m *mocks.UserRepo) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   map[string]string{"error": "invalid token subclaims"},
		},
		{
			name:       "User not found",
			authHeader: "VALID",
			mockSetup: func(m *mocks.UserRepo) {
				m.On("GetByID", mock.Anything, 1).Return(nil, nil).Once()
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   map[string]string{"error": "user not found"},
		},
		{
			name:       "Database error during GetByID",
			authHeader: "VALID",
			mockSetup: func(m *mocks.UserRepo) {
				m.On("GetByID", mock.Anything, 1).Return(nil, errors.New("database error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]string{"error": "failed to get user by ID"},
		},
	}

	secret := "your-secret-key"
	e := echo.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(mocks.UserRepo)
			tt.mockSetup(mockUserRepo)

			authMiddleware := NewAuthMiddleware(mockUserRepo, secret)

			var header string
			switch tt.authHeader {
			case "VALID":
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": 1})
				tokenString, err := token.SignedString([]byte(secret))
				if err != nil {
					t.Fatalf("failed to sign token: %v", err)
				}
				header = "Bearer " + tokenString
			case "INVALID_CLAIMS":
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "not_a_number"})
				tokenString, err := token.SignedString([]byte(secret))
				if err != nil {
					t.Fatalf("failed to sign token: %v", err)
				}
				header = "Bearer " + tokenString
			default:
				header = tt.authHeader
			}

			req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
			req.Header.Set("Authorization", header)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			nextHandler := func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			}

			wrappedHandler := authMiddleware(nextHandler)
			err := wrappedHandler(c)

			if tt.expectedBody == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
			} else {
				assert.Equal(t, tt.expectedStatus, rec.Code)
				var body map[string]string
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}
				assert.Equal(t, tt.expectedBody, body)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gratefultolord/merch-store/internal/repository"
	"github.com/labstack/echo/v4"
)

func NewAuthMiddleware(userRepo repository.UserRepo, secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid authorization header format"})
			}

			tokenString := parts[1]

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(secret), nil
			})

			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				subClaim, ok := claims["sub"]
				if !ok {
					return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
				}

				userID, ok := subClaim.(float64)
				if !ok {
					return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token subclaims"})
				}
				userIDInt := int(userID)

				user, err := userRepo.GetByID(context.Background(), userIDInt)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get user by ID"})
				}
				if user == nil {
					return c.JSON(http.StatusUnauthorized, map[string]string{"error": "user not found"})
				}

				c.Set("userID", userIDInt)

				return next(c)
			}

			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
		}
	}
}

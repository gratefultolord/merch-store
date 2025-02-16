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

// NewAuthMiddleware создает middleware для проверки JWT-токена
func NewAuthMiddleware(userRepo repository.UserRepo, secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Получаем значение заголовка Authorization
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "mw: missing authorization header"})
			}

			// Разделяем заголовок на две части: "Bearer" и сам токен
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "mw: invalid authorization header format"})
			}

			// Получаем сам JWT-токен
			tokenString := parts[1]

			// Парсим JWT-токен
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Проверяем метод подписи токена
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				// Возвращаем секретный ключ для проверки подписи
				return []byte(secret), nil
			})

			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "mw: invalid token 1"})
			}

			// Проверяем, является ли токен действительным
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				// Получаем идентификатор пользователя из клеймов токена
				subClaim, ok := claims["sub"]
				if !ok {
					return c.JSON(http.StatusUnauthorized, map[string]string{"error": "mw: invalid token claims"})
				}

				// Преобразуем идентификатор пользователя в целое число
				userID, ok := subClaim.(float64)
				if !ok {
					return c.JSON(http.StatusUnauthorized, map[string]string{"error": "mw: invalid token subclaims"})
				}
				userIDInt := int(userID)

				// Получаем пользователя по идентификатору
				user, err := userRepo.GetByID(context.Background(), userIDInt)
				if err != nil {
					fmt.Printf("mw: userRepo.GetByID err: %v", err)
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "mw: failed to get user by ID"})
				}
				if user == nil {
					return c.JSON(http.StatusUnauthorized, map[string]string{"error": "mw: user not found"})
				}

				// Добавляем идентификатор пользователя в контекст запроса
				c.Set("userID", userIDInt)

				// Вызываем следующий обработчик
				return next(c)
			}

			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "mw: invalid token 2"})
		}
	}
}

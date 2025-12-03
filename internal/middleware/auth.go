package middleware

import (
	"net/http"
	"strings"

	"github.com/auth-service/internal/utils"
	"github.com/labstack/echo/v4"
)

type AuthMiddleware struct {
	jwtManager *utils.JWTManager
}

func NewAuthMiddleware(jwtManager *utils.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager}
}

func (m *AuthMiddleware) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error": map[string]string{
					"code":    "MISSING_TOKEN",
					"message": "Authorization header is required",
				},
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error": map[string]string{
					"code":    "INVALID_TOKEN_FORMAT",
					"message": "Invalid authorization header format",
				},
			})
		}

		claims, err := m.jwtManager.ValidateToken(parts[1])
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error": map[string]string{
					"code":    "INVALID_TOKEN",
					"message": "Invalid or expired token",
				},
			})
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)

		return next(c)
	}
}

func (m *AuthMiddleware) RequireRoles(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRoles, ok := c.Get("roles").([]string)
			if !ok {
				return c.JSON(http.StatusForbidden, map[string]interface{}{
					"error": map[string]string{
						"code":    "FORBIDDEN",
						"message": "Access denied",
					},
				})
			}

			for _, requiredRole := range roles {
				for _, userRole := range userRoles {
					if userRole == requiredRole {
						return next(c)
					}
				}
			}

			return c.JSON(http.StatusForbidden, map[string]interface{}{
				"error": map[string]string{
					"code":    "FORBIDDEN",
					"message": "Insufficient permissions",
				},
			})
		}
	}
}

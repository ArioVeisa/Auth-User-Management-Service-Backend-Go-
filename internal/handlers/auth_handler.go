package handlers

import (
	"net/http"

	"github.com/auth-service/internal/models"
	"github.com/auth-service/internal/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req models.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
			},
		})
	}

	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "VALIDATION_ERROR",
				"message": "Email, password, and display_name are required",
			},
		})
	}

	ip := c.RealIP()
	userAgent := c.Request().UserAgent()

	user, err := h.authService.Register(req, ip, userAgent)
	if err != nil {
		if err == services.ErrDuplicateEmail {
			return c.JSON(http.StatusConflict, map[string]interface{}{
				"error": map[string]string{
					"code":    "DUPLICATE_EMAIL",
					"message": "Email already exists",
				},
			})
		}
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "REGISTRATION_FAILED",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Registration successful. Please check your email to verify your account.",
		"user": map[string]interface{}{
			"id":           user.ID,
			"email":        user.Email,
			"display_name": user.DisplayName,
		},
	})
}

func (h *AuthHandler) VerifyEmail(c echo.Context) error {
	var req models.VerifyEmailRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
			},
		})
	}

	if req.Token == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "VALIDATION_ERROR",
				"message": "Token is required",
			},
		})
	}

	ip := c.RealIP()
	userAgent := c.Request().UserAgent()

	if err := h.authService.VerifyEmail(req.Token, ip, userAgent); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "VERIFICATION_FAILED",
				"message": "Invalid or expired token",
			},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Email verified successfully",
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
			},
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "VALIDATION_ERROR",
				"message": "Email and password are required",
			},
		})
	}

	ip := c.RealIP()
	userAgent := c.Request().UserAgent()

	response, err := h.authService.Login(req, ip, userAgent)
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error": map[string]string{
					"code":    "INVALID_CREDENTIALS",
					"message": "Invalid email or password",
				},
			})
		case services.ErrUserNotVerified:
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error": map[string]string{
					"code":    "EMAIL_NOT_VERIFIED",
					"message": "Please verify your email first",
				},
			})
		case services.ErrAccountLocked:
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error": map[string]string{
					"code":    "ACCOUNT_LOCKED",
					"message": "Account is temporarily locked due to too many failed attempts",
				},
			})
		default:
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error": map[string]string{
					"code":    "LOGIN_FAILED",
					"message": err.Error(),
				},
			})
		}
	}

	return c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var req models.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
			},
		})
	}

	if req.RefreshToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "VALIDATION_ERROR",
				"message": "Refresh token is required",
			},
		})
	}

	ip := c.RealIP()
	userAgent := c.Request().UserAgent()

	response, err := h.authService.RefreshToken(req.RefreshToken, ip, userAgent)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": map[string]string{
				"code":    "TOKEN_REFRESH_FAILED",
				"message": "Invalid or expired refresh token",
			},
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": map[string]string{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
	}

	ip := c.RealIP()
	userAgent := c.Request().UserAgent()

	if err := h.authService.Logout(userID, ip, userAgent); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": map[string]string{
				"code":    "LOGOUT_FAILED",
				"message": "Failed to logout",
			},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Logged out successfully",
	})
}

func (h *AuthHandler) ForgotPassword(c echo.Context) error {
	var req models.ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
			},
		})
	}

	ip := c.RealIP()
	userAgent := c.Request().UserAgent()

	h.authService.ForgotPassword(req.Email, ip, userAgent)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "If the email exists, a password reset link will be sent",
	})
}

func (h *AuthHandler) ResetPassword(c echo.Context) error {
	var req models.ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
			},
		})
	}

	if req.Token == "" || req.NewPassword == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "VALIDATION_ERROR",
				"message": "Token and new_password are required",
			},
		})
	}

	ip := c.RealIP()
	userAgent := c.Request().UserAgent()

	if err := h.authService.ResetPassword(req.Token, req.NewPassword, ip, userAgent); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "RESET_FAILED",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Password reset successfully",
	})
}

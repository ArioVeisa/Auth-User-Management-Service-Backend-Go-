package handlers

import (
	"net/http"
	"strconv"

	"github.com/auth-service/internal/models"
	"github.com/auth-service/internal/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetUser(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_ID",
				"message": "Invalid user ID format",
			},
		})
	}

	user, err := h.userService.GetUser(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": map[string]string{
				"code":    "USER_NOT_FOUND",
				"message": "User not found",
			},
		})
	}

	return c.JSON(http.StatusOK, user)
}

func (h *UserHandler) ListUsers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	search := c.QueryParam("search")

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	result, err := h.userService.ListUsers(page, perPage, search)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": map[string]string{
				"code":    "LIST_FAILED",
				"message": "Failed to list users",
			},
		})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *UserHandler) CreateUser(c echo.Context) error {
	var req models.CreateUserRequest
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

	createdBy, _ := c.Get("user_id").(uuid.UUID)
	ip := c.RealIP()
	userAgent := c.Request().UserAgent()

	user, err := h.userService.CreateUser(req, createdBy, ip, userAgent)
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
				"code":    "CREATE_FAILED",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) UpdateUser(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_ID",
				"message": "Invalid user ID format",
			},
		})
	}

	var req models.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
			},
		})
	}

	updatedBy, _ := c.Get("user_id").(uuid.UUID)
	ip := c.RealIP()
	userAgent := c.Request().UserAgent()

	user, err := h.userService.UpdateUser(id, req, updatedBy, ip, userAgent)
	if err != nil {
		if err == services.ErrUserNotFound {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": map[string]string{
					"code":    "USER_NOT_FOUND",
					"message": "User not found",
				},
			})
		}
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "UPDATE_FAILED",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_ID",
				"message": "Invalid user ID format",
			},
		})
	}

	if err := h.userService.DeleteUser(id); err != nil {
		if err == services.ErrUserNotFound {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error": map[string]string{
					"code":    "USER_NOT_FOUND",
					"message": "User not found",
				},
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": map[string]string{
				"code":    "DELETE_FAILED",
				"message": "Failed to delete user",
			},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "User deleted successfully",
	})
}

func (h *UserHandler) AssignRole(c echo.Context) error {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_ID",
				"message": "Invalid user ID format",
			},
		})
	}

	var req models.AssignRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
			},
		})
	}

	assignedBy, _ := c.Get("user_id").(uuid.UUID)
	ip := c.RealIP()
	userAgent := c.Request().UserAgent()

	if err := h.userService.AssignRole(userID, req.RoleID, assignedBy, ip, userAgent); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "ASSIGN_ROLE_FAILED",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Role assigned successfully",
	})
}

func (h *UserHandler) UnassignRole(c echo.Context) error {
	idStr := c.Param("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_ID",
				"message": "Invalid user ID format",
			},
		})
	}

	roleIDStr := c.Param("role")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_ROLE_ID",
				"message": "Invalid role ID format",
			},
		})
	}

	removedBy, _ := c.Get("user_id").(uuid.UUID)
	ip := c.RealIP()
	userAgent := c.Request().UserAgent()

	if err := h.userService.UnassignRole(userID, roleID, removedBy, ip, userAgent); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "UNASSIGN_ROLE_FAILED",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Role unassigned successfully",
	})
}

func (h *UserHandler) GetCurrentUser(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": map[string]string{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
	}

	user, err := h.userService.GetUser(userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": map[string]string{
				"code":    "USER_NOT_FOUND",
				"message": "User not found",
			},
		})
	}

	return c.JSON(http.StatusOK, user)
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (h *UserHandler) ChangePassword(c echo.Context) error {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": map[string]string{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
	}

	var req ChangePasswordRequest
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

	if err := h.userService.ChangePassword(userID, req.OldPassword, req.NewPassword, ip, userAgent); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "CHANGE_PASSWORD_FAILED",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Password changed successfully",
	})
}

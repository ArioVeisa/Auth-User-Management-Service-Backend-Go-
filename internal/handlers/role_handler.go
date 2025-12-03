package handlers

import (
	"net/http"
	"strconv"

	"github.com/auth-service/internal/models"
	"github.com/auth-service/internal/services"
	"github.com/labstack/echo/v4"
)

type RoleHandler struct {
	roleService *services.RoleService
}

func NewRoleHandler(roleService *services.RoleService) *RoleHandler {
	return &RoleHandler{roleService: roleService}
}

func (h *RoleHandler) GetRole(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_ID",
				"message": "Invalid role ID format",
			},
		})
	}

	role, err := h.roleService.GetRole(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error": map[string]string{
				"code":    "ROLE_NOT_FOUND",
				"message": "Role not found",
			},
		})
	}

	return c.JSON(http.StatusOK, role)
}

func (h *RoleHandler) ListRoles(c echo.Context) error {
	roles, err := h.roleService.ListRoles()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": map[string]string{
				"code":    "LIST_FAILED",
				"message": "Failed to list roles",
			},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": roles,
	})
}

func (h *RoleHandler) CreateRole(c echo.Context) error {
	var req models.CreateRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
			},
		})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "VALIDATION_ERROR",
				"message": "Role name is required",
			},
		})
	}

	role, err := h.roleService.CreateRole(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "CREATE_FAILED",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(http.StatusCreated, role)
}

func (h *RoleHandler) UpdateRole(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_ID",
				"message": "Invalid role ID format",
			},
		})
	}

	var req models.CreateRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request body",
			},
		})
	}

	role, err := h.roleService.UpdateRole(id, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "UPDATE_FAILED",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(http.StatusOK, role)
}

func (h *RoleHandler) DeleteRole(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "INVALID_ID",
				"message": "Invalid role ID format",
			},
		})
	}

	if err := h.roleService.DeleteRole(id); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": map[string]string{
				"code":    "DELETE_FAILED",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Role deleted successfully",
	})
}

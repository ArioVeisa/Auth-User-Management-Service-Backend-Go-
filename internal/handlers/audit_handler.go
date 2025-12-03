package handlers

import (
	"net/http"
	"strconv"

	"github.com/auth-service/internal/models"
	"github.com/auth-service/internal/services"
	"github.com/labstack/echo/v4"
)

type AuditHandler struct {
	auditService *services.AuditService
}

func NewAuditHandler(auditService *services.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

func (h *AuditHandler) ListAuditLogs(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	query := models.AuditQuery{
		UserID:    c.QueryParam("user_id"),
		EventType: c.QueryParam("event_type"),
		StartDate: c.QueryParam("start_date"),
		EndDate:   c.QueryParam("end_date"),
		Page:      page,
		PerPage:   perPage,
	}

	events, total, err := h.auditService.GetAuditLogs(query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": map[string]string{
				"code":    "LIST_FAILED",
				"message": "Failed to list audit logs",
			},
		})
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       events,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	})
}

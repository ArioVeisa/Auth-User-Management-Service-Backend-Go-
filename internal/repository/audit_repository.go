package repository

import (
	"database/sql"
	"time"

	"github.com/auth-service/internal/models"
	"github.com/google/uuid"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(event *models.AuditEvent) error {
	query := `
		INSERT INTO audit_events (id, user_id, event_type, payload, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(query, event.ID, event.UserID, event.EventType,
		event.Payload, event.IPAddress, event.UserAgent, event.CreatedAt)
	return err
}

func (r *AuditRepository) List(query models.AuditQuery) ([]models.AuditEvent, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 {
		query.PerPage = 20
	}
	offset := (query.Page - 1) * query.PerPage

	baseQuery := "FROM audit_events WHERE 1=1"
	args := []interface{}{}
	argCount := 0

	if query.UserID != "" {
		argCount++
		baseQuery += " AND user_id = $" + string(rune('0'+argCount))
		userID, _ := uuid.Parse(query.UserID)
		args = append(args, userID)
	}

	if query.EventType != "" {
		argCount++
		baseQuery += " AND event_type = $" + string(rune('0'+argCount))
		args = append(args, query.EventType)
	}

	if query.StartDate != "" {
		argCount++
		startDate, _ := time.Parse("2006-01-02", query.StartDate)
		baseQuery += " AND created_at >= $" + string(rune('0'+argCount))
		args = append(args, startDate)
	}

	if query.EndDate != "" {
		argCount++
		endDate, _ := time.Parse("2006-01-02", query.EndDate)
		baseQuery += " AND created_at <= $" + string(rune('0'+argCount))
		args = append(args, endDate.Add(24*time.Hour))
	}

	var total int64
	countSQL := "SELECT COUNT(*) " + baseQuery
	r.db.QueryRow(countSQL, args...).Scan(&total)

	selectSQL := "SELECT id, user_id, event_type, payload, ip_address, user_agent, created_at " +
		baseQuery + " ORDER BY created_at DESC LIMIT $" + string(rune('0'+argCount+1)) +
		" OFFSET $" + string(rune('0'+argCount+2))
	args = append(args, query.PerPage, offset)

	rows, err := r.db.Query(selectSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []models.AuditEvent
	for rows.Next() {
		var event models.AuditEvent
		if err := rows.Scan(&event.ID, &event.UserID, &event.EventType,
			&event.Payload, &event.IPAddress, &event.UserAgent, &event.CreatedAt); err != nil {
			return nil, 0, err
		}
		events = append(events, event)
	}

	return events, total, nil
}

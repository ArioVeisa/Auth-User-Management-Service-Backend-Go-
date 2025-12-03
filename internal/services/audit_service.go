package services

import (
	"encoding/json"
	"time"

	"github.com/auth-service/internal/models"
	"github.com/auth-service/internal/repository"
	"github.com/google/uuid"
)

type AuditService struct {
	auditRepo *repository.AuditRepository
}

func NewAuditService(auditRepo *repository.AuditRepository) *AuditService {
	return &AuditService{auditRepo: auditRepo}
}

func (s *AuditService) LogEvent(eventType models.AuditEventType, userID *uuid.UUID, payload map[string]interface{}, ip, userAgent string) error {
	payloadJSON, _ := json.Marshal(payload)
	
	event := &models.AuditEvent{
		ID:        uuid.New(),
		UserID:    userID,
		EventType: eventType,
		Payload:   string(payloadJSON),
		IPAddress: ip,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	}

	return s.auditRepo.Create(event)
}

func (s *AuditService) GetAuditLogs(query models.AuditQuery) ([]models.AuditEvent, int64, error) {
	return s.auditRepo.List(query)
}

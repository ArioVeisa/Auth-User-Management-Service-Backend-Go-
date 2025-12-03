package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID               uuid.UUID  `json:"id"`
	Email            string     `json:"email"`
	PasswordHash     string     `json:"-"`
	DisplayName      string     `json:"display_name"`
	IsActive         bool       `json:"is_active"`
	IsVerified       bool       `json:"is_verified"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty"`
	FailedLoginCount int        `json:"-"`
	LockedUntil      *time.Time `json:"-"`
}

type Role struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UserRole struct {
	UserID     uuid.UUID `json:"user_id"`
	RoleID     int       `json:"role_id"`
	AssignedBy uuid.UUID `json:"assigned_by"`
	AssignedAt time.Time `json:"assigned_at"`
}

type RefreshToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"-"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
	UserAgent string    `json:"user_agent"`
	IPAddress string    `json:"ip_address"`
}

type EmailTokenType string

const (
	EmailTokenTypeVerify EmailTokenType = "verify"
	EmailTokenTypeReset  EmailTokenType = "reset"
)

type EmailToken struct {
	ID        uuid.UUID      `json:"id"`
	UserID    uuid.UUID      `json:"user_id"`
	TokenHash string         `json:"-"`
	Type      EmailTokenType `json:"type"`
	ExpiresAt time.Time      `json:"expires_at"`
	Used      bool           `json:"used"`
}

type AuditEventType string

const (
	AuditEventLoginSuccess   AuditEventType = "login_success"
	AuditEventLoginFailed    AuditEventType = "login_failed"
	AuditEventLogout         AuditEventType = "logout"
	AuditEventPasswordChange AuditEventType = "password_change"
	AuditEventRoleChange     AuditEventType = "role_change"
	AuditEventRegister       AuditEventType = "register"
	AuditEventEmailVerified  AuditEventType = "email_verified"
	AuditEventPasswordReset  AuditEventType = "password_reset"
)

type AuditEvent struct {
	ID        uuid.UUID      `json:"id"`
	UserID    *uuid.UUID     `json:"user_id,omitempty"`
	EventType AuditEventType `json:"event_type"`
	Payload   string         `json:"payload"`
	IPAddress string         `json:"ip_address"`
	UserAgent string         `json:"user_agent"`
	CreatedAt time.Time      `json:"created_at"`
}

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type VerifyEmailRequest struct {
	Token string `json:"token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type UserWithRoles struct {
	User  User   `json:"user"`
	Roles []Role `json:"roles"`
}

type CreateUserRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	RoleIDs     []int  `json:"role_ids,omitempty"`
}

type UpdateUserRequest struct {
	Email       *string `json:"email,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

type CreateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AssignRoleRequest struct {
	RoleID int `json:"role_id"`
}

type PaginationQuery struct {
	Page    int    `query:"page"`
	PerPage int    `query:"per_page"`
	Search  string `query:"search"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"total_pages"`
}

type AuditQuery struct {
	UserID    string `query:"user_id"`
	EventType string `query:"event_type"`
	StartDate string `query:"start_date"`
	EndDate   string `query:"end_date"`
	Page      int    `query:"page"`
	PerPage   int    `query:"per_page"`
}

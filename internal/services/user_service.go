package services

import (
	"errors"
	"time"

	"github.com/auth-service/internal/models"
	"github.com/auth-service/internal/repository"
	"github.com/auth-service/internal/utils"
	"github.com/google/uuid"
)

type UserService struct {
	userRepo     *repository.UserRepository
	roleRepo     *repository.RoleRepository
	auditService *AuditService
}

func NewUserService(
	userRepo *repository.UserRepository,
	roleRepo *repository.RoleRepository,
	auditService *AuditService,
) *UserService {
	return &UserService{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		auditService: auditService,
	}
}

func (s *UserService) GetUser(id uuid.UUID) (*models.UserWithRoles, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	roles, _ := s.roleRepo.GetUserRoles(id)

	return &models.UserWithRoles{
		User:  *user,
		Roles: roles,
	}, nil
}

func (s *UserService) ListUsers(page, perPage int, search string) (*models.PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	users, total, err := s.userRepo.List(page, perPage, search)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return &models.PaginatedResponse{
		Data:       users,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *UserService) CreateUser(req models.CreateUserRequest, createdBy uuid.UUID, ip, userAgent string) (*models.User, error) {
	email := utils.SanitizeEmail(req.Email)

	if !utils.ValidateEmail(email) {
		return nil, errors.New("invalid email format")
	}

	if valid, msg := utils.ValidatePassword(req.Password); !valid {
		return nil, errors.New(msg)
	}

	existingUser, _ := s.userRepo.GetByEmail(email)
	if existingUser != nil {
		return nil, ErrDuplicateEmail
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  req.DisplayName,
		IsActive:     true,
		IsVerified:   true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	for _, roleID := range req.RoleIDs {
		s.roleRepo.AssignRoleToUser(user.ID, roleID, createdBy)
	}

	return user, nil
}

func (s *UserService) UpdateUser(id uuid.UUID, req models.UpdateUserRequest, updatedBy uuid.UUID, ip, userAgent string) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if req.Email != nil {
		email := utils.SanitizeEmail(*req.Email)
		if !utils.ValidateEmail(email) {
			return nil, errors.New("invalid email format")
		}
		existingUser, _ := s.userRepo.GetByEmail(email)
		if existingUser != nil && existingUser.ID != id {
			return nil, ErrDuplicateEmail
		}
		user.Email = email
	}

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) DeleteUser(id uuid.UUID) error {
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		return ErrUserNotFound
	}

	return s.userRepo.Delete(id)
}

func (s *UserService) AssignRole(userID uuid.UUID, roleID int, assignedBy uuid.UUID, ip, userAgent string) error {
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	_, err = s.roleRepo.GetByID(roleID)
	if err != nil {
		return errors.New("role not found")
	}

	if err := s.roleRepo.AssignRoleToUser(userID, roleID, assignedBy); err != nil {
		return err
	}

	s.auditService.LogEvent(models.AuditEventRoleChange, &userID, map[string]interface{}{
		"action":      "assign",
		"role_id":     roleID,
		"assigned_by": assignedBy.String(),
	}, ip, userAgent)

	return nil
}

func (s *UserService) UnassignRole(userID uuid.UUID, roleID int, removedBy uuid.UUID, ip, userAgent string) error {
	if err := s.roleRepo.UnassignRoleFromUser(userID, roleID); err != nil {
		return err
	}

	s.auditService.LogEvent(models.AuditEventRoleChange, &userID, map[string]interface{}{
		"action":     "unassign",
		"role_id":    roleID,
		"removed_by": removedBy.String(),
	}, ip, userAgent)

	return nil
}

func (s *UserService) ChangePassword(userID uuid.UUID, oldPassword, newPassword, ip, userAgent string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	if !utils.CheckPassword(oldPassword, user.PasswordHash) {
		return ErrInvalidCredentials
	}

	if valid, msg := utils.ValidatePassword(newPassword); !valid {
		return errors.New(msg)
	}

	passwordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(userID, passwordHash); err != nil {
		return err
	}

	s.auditService.LogEvent(models.AuditEventPasswordChange, &userID, nil, ip, userAgent)

	return nil
}

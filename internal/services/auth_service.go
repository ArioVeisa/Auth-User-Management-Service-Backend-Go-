package services

import (
	"database/sql"
	"errors"
	"time"

	"github.com/auth-service/internal/config"
	"github.com/auth-service/internal/models"
	"github.com/auth-service/internal/repository"
	"github.com/auth-service/internal/utils"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotVerified    = errors.New("email not verified")
	ErrAccountLocked      = errors.New("account is locked")
	ErrDuplicateEmail     = errors.New("email already exists")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrTokenRevoked       = errors.New("token has been revoked")
)

type AuthService struct {
	cfg          *config.Config
	userRepo     *repository.UserRepository
	tokenRepo    *repository.TokenRepository
	roleRepo     *repository.RoleRepository
	emailService *EmailService
	auditService *AuditService
	jwtManager   *utils.JWTManager
}

func NewAuthService(
	cfg *config.Config,
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	roleRepo *repository.RoleRepository,
	emailService *EmailService,
	auditService *AuditService,
) *AuthService {
	return &AuthService{
		cfg:          cfg,
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		roleRepo:     roleRepo,
		emailService: emailService,
		auditService: auditService,
		jwtManager:   utils.NewJWTManager(cfg.JWTSigningKey, cfg.AccessTokenExpiry),
	}
}

func (s *AuthService) Register(req models.RegisterRequest, ip, userAgent string) (*models.User, error) {
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
		IsVerified:   false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	defaultRole, _ := s.roleRepo.GetByName("user")
	if defaultRole != nil {
		s.roleRepo.AssignRoleToUser(user.ID, defaultRole.ID, user.ID)
	}

	token, _ := utils.GenerateRandomToken(32)
	emailToken := &models.EmailToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: utils.HashToken(token),
		Type:      models.EmailTokenTypeVerify,
		ExpiresAt: time.Now().Add(time.Hour),
		Used:      false,
	}
	s.tokenRepo.CreateEmailToken(emailToken)

	go s.emailService.SendVerificationEmail(user.Email, user.DisplayName, token)

	s.auditService.LogEvent(models.AuditEventRegister, &user.ID, map[string]interface{}{
		"email": user.Email,
	}, ip, userAgent)

	return user, nil
}

func (s *AuthService) VerifyEmail(tokenStr, ip, userAgent string) error {
	tokenHash := utils.HashToken(tokenStr)
	emailToken, err := s.tokenRepo.GetEmailTokenByHash(tokenHash, models.EmailTokenTypeVerify)
	if err != nil || emailToken == nil {
		return ErrInvalidToken
	}

	if emailToken.Used || time.Now().After(emailToken.ExpiresAt) {
		return ErrInvalidToken
	}

	if err := s.tokenRepo.MarkEmailTokenUsed(emailToken.ID); err != nil {
		return err
	}

	if err := s.userRepo.SetVerified(emailToken.UserID); err != nil {
		return err
	}

	s.auditService.LogEvent(models.AuditEventEmailVerified, &emailToken.UserID, nil, ip, userAgent)

	return nil
}

func (s *AuthService) Login(req models.LoginRequest, ip, userAgent string) (*models.AuthResponse, error) {
	email := utils.SanitizeEmail(req.Email)

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		s.auditService.LogEvent(models.AuditEventLoginFailed, &user.ID, map[string]interface{}{
			"reason": "account_locked",
		}, ip, userAgent)
		return nil, ErrAccountLocked
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		s.userRepo.IncrementFailedLogin(user.ID)

		if user.FailedLoginCount+1 >= s.cfg.MaxFailedLogins {
			lockUntil := time.Now().Add(s.cfg.LockDuration)
			s.userRepo.LockAccount(user.ID, lockUntil)
		}

		s.auditService.LogEvent(models.AuditEventLoginFailed, &user.ID, map[string]interface{}{
			"reason": "invalid_password",
		}, ip, userAgent)
		return nil, ErrInvalidCredentials
	}

	if !user.IsVerified {
		return nil, ErrUserNotVerified
	}

	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	s.userRepo.ResetFailedLogin(user.ID)

	roles, _ := s.roleRepo.GetUserRoles(user.ID)
	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name
	}

	accessToken, err := s.jwtManager.GenerateToken(user.ID, user.Email, roleNames)
	if err != nil {
		return nil, err
	}

	refreshTokenStr, _ := utils.GenerateRandomToken(32)
	refreshToken := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: utils.HashToken(refreshTokenStr),
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenExpiry),
		Revoked:   false,
		UserAgent: userAgent,
		IPAddress: ip,
	}

	if err := s.tokenRepo.CreateRefreshToken(refreshToken); err != nil {
		return nil, err
	}

	s.auditService.LogEvent(models.AuditEventLoginSuccess, &user.ID, nil, ip, userAgent)

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		ExpiresIn:    int64(s.cfg.AccessTokenExpiry.Seconds()),
	}, nil
}

func (s *AuthService) RefreshToken(tokenStr, ip, userAgent string) (*models.AuthResponse, error) {
	tokenHash := utils.HashToken(tokenStr)
	oldToken, err := s.tokenRepo.GetRefreshTokenByHash(tokenHash)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if oldToken.Revoked {
		s.tokenRepo.RevokeAllUserTokens(oldToken.UserID)
		return nil, ErrTokenRevoked
	}

	if time.Now().After(oldToken.ExpiresAt) {
		return nil, ErrInvalidToken
	}

	s.tokenRepo.RevokeRefreshToken(oldToken.ID)

	user, err := s.userRepo.GetByID(oldToken.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	roles, _ := s.roleRepo.GetUserRoles(user.ID)
	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name
	}

	accessToken, err := s.jwtManager.GenerateToken(user.ID, user.Email, roleNames)
	if err != nil {
		return nil, err
	}

	newRefreshTokenStr, _ := utils.GenerateRandomToken(32)
	newRefreshToken := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: utils.HashToken(newRefreshTokenStr),
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(s.cfg.RefreshTokenExpiry),
		Revoked:   false,
		UserAgent: userAgent,
		IPAddress: ip,
	}

	if err := s.tokenRepo.CreateRefreshToken(newRefreshToken); err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshTokenStr,
		ExpiresIn:    int64(s.cfg.AccessTokenExpiry.Seconds()),
	}, nil
}

func (s *AuthService) Logout(userID uuid.UUID, ip, userAgent string) error {
	if err := s.tokenRepo.RevokeAllUserTokens(userID); err != nil {
		return err
	}

	s.auditService.LogEvent(models.AuditEventLogout, &userID, nil, ip, userAgent)

	return nil
}

func (s *AuthService) ForgotPassword(email, ip, userAgent string) error {
	email = utils.SanitizeEmail(email)
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil
	}

	token, _ := utils.GenerateRandomToken(32)
	emailToken := &models.EmailToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: utils.HashToken(token),
		Type:      models.EmailTokenTypeReset,
		ExpiresAt: time.Now().Add(time.Hour),
		Used:      false,
	}
	s.tokenRepo.CreateEmailToken(emailToken)

	go s.emailService.SendPasswordResetEmail(user.Email, user.DisplayName, token)

	return nil
}

func (s *AuthService) ResetPassword(tokenStr, newPassword, ip, userAgent string) error {
	if valid, msg := utils.ValidatePassword(newPassword); !valid {
		return errors.New(msg)
	}

	tokenHash := utils.HashToken(tokenStr)
	emailToken, err := s.tokenRepo.GetEmailTokenByHash(tokenHash, models.EmailTokenTypeReset)
	if err != nil || emailToken == nil {
		return ErrInvalidToken
	}

	if emailToken.Used || time.Now().After(emailToken.ExpiresAt) {
		return ErrInvalidToken
	}

	passwordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(emailToken.UserID, passwordHash); err != nil {
		return err
	}

	s.tokenRepo.MarkEmailTokenUsed(emailToken.ID)
	s.tokenRepo.RevokeAllUserTokens(emailToken.UserID)

	s.auditService.LogEvent(models.AuditEventPasswordReset, &emailToken.UserID, nil, ip, userAgent)

	return nil
}

func (s *AuthService) ValidateAccessToken(tokenStr string) (*utils.JWTClaims, error) {
	return s.jwtManager.ValidateToken(tokenStr)
}

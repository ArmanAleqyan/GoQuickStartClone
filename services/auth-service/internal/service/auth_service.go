package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"ironnode/pkg/email"
	"ironnode/pkg/models"
	"ironnode/services/auth-service/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(email, password, firstName, lastName string) (*models.User, error)
	Login(email, password string) (string, error)
	ValidateToken(tokenString string) (*uuid.UUID, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	ForgotPassword(email string) (string, error)
	VerifyResetToken(token string) (*models.PasswordReset, error)
	ResetPassword(token, newPassword string) error
}

type authService struct {
	repo         repository.AuthRepository
	emailService *email.EmailService
	jwtSecret    string
	jwtExpiry    time.Duration
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

func NewAuthService(repo repository.AuthRepository, emailService *email.EmailService, jwtSecret string, jwtExpiry time.Duration) AuthService {
	return &authService{
		repo:         repo,
		emailService: emailService,
		jwtSecret:    jwtSecret,
		jwtExpiry:    jwtExpiry,
	}
}

func (s *authService) Register(email, password, firstName, lastName string) (*models.User, error) {
	// Check if user already exists
	existingUser, _ := s.repo.GetUserByEmail(email)
	if existingUser != nil {
		return nil, errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:     email,
		Password:  string(hashedPassword),
		FirstName: firstName,
		LastName:  lastName,
		IsActive:  true,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(email, password string) (string, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if !user.IsActive {
		return "", errors.New("user account is inactive")
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *authService) ValidateToken(tokenString string) (*uuid.UUID, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return &claims.UserID, nil
}

func (s *authService) GetUserByID(id uuid.UUID) (*models.User, error) {
	return s.repo.GetUserByID(id)
}

func (s *authService) ForgotPassword(email string) (string, error) {
	// Check if user exists
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		// For security, don't reveal if user exists or not
		return "", nil
	}

	// Invalidate all existing password reset tokens for this user
	if err := s.repo.InvalidateUserPasswordResets(user.ID); err != nil {
		return "", err
	}

	// Generate secure random token
	token, err := generateSecureToken(32)
	if err != nil {
		return "", err
	}

	// Create password reset record
	passwordReset := &models.PasswordReset{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour), // Token valid for 1 hour
	}

	if err := s.repo.CreatePasswordReset(passwordReset); err != nil {
		return "", err
	}

	// Send password reset email
	resetURL := "http://localhost:3000/reset-password" // TODO: Get from config
	if s.emailService != nil {
		if err := s.emailService.SendPasswordResetEmail(user.Email, token, resetURL); err != nil {
			// Log error but don't fail the request
			return token, nil
		}
	}

	return token, nil
}

func (s *authService) VerifyResetToken(token string) (*models.PasswordReset, error) {
	reset, err := s.repo.GetPasswordResetByToken(token)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	if !reset.IsValid() {
		return nil, errors.New("invalid or expired token")
	}

	return reset, nil
}

func (s *authService) ResetPassword(token, newPassword string) error {
	// Verify token
	reset, err := s.VerifyResetToken(token)
	if err != nil {
		return err
	}

	// Get user
	user, err := s.repo.GetUserByID(reset.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update user password
	user.Password = string(hashedPassword)
	if err := s.repo.UpdateUser(user); err != nil {
		return err
	}

	// Mark token as used
	if err := s.repo.MarkPasswordResetAsUsed(reset.ID); err != nil {
		return err
	}

	// Send confirmation email
	if s.emailService != nil {
		s.emailService.SendPasswordChangedEmail(user.Email)
	}

	return nil
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

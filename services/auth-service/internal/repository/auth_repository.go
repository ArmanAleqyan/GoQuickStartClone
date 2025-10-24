package repository

import (
	"errors"

	"ironnode/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthRepository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	UpdateUser(user *models.User) error
	CreatePasswordReset(reset *models.PasswordReset) error
	GetPasswordResetByToken(token string) (*models.PasswordReset, error)
	InvalidateUserPasswordResets(userID uuid.UUID) error
	MarkPasswordResetAsUsed(resetID uuid.UUID) error
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *authRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *authRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *authRepository) CreatePasswordReset(reset *models.PasswordReset) error {
	return r.db.Create(reset).Error
}

func (r *authRepository) GetPasswordResetByToken(token string) (*models.PasswordReset, error) {
	var reset models.PasswordReset
	err := r.db.Preload("User").Where("token = ?", token).First(&reset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("password reset token not found")
		}
		return nil, err
	}
	return &reset, nil
}

func (r *authRepository) InvalidateUserPasswordResets(userID uuid.UUID) error {
	// Mark all existing unused password resets for this user as used
	now := gorm.Expr("NOW()")
	return r.db.Model(&models.PasswordReset{}).
		Where("user_id = ? AND used_at IS NULL", userID).
		Update("used_at", now).Error
}

func (r *authRepository) MarkPasswordResetAsUsed(resetID uuid.UUID) error {
	now := gorm.Expr("NOW()")
	return r.db.Model(&models.PasswordReset{}).
		Where("id = ?", resetID).
		Update("used_at", now).Error
}

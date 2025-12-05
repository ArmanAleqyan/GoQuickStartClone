package repository

import (
	"ironnode/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WalletRepository interface {
	CreateWallet(wallet *models.Wallet) error
	GetWalletByID(id uuid.UUID) (*models.Wallet, error)
	GetWalletsByUserID(userID uuid.UUID) ([]*models.Wallet, error)
	GetWalletsByClientUserID(clientUserID string) ([]*models.Wallet, error)
	GetWalletByAddress(address string) (*models.Wallet, error)
	GetWalletsByUserAndClient(userID uuid.UUID, clientUserID string) ([]*models.Wallet, error)
	UpdateWallet(wallet *models.Wallet) error
	DeleteWallet(id uuid.UUID) error
	DeactivateWallet(id uuid.UUID) error
}

type walletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) CreateWallet(wallet *models.Wallet) error {
	return r.db.Create(wallet).Error
}

func (r *walletRepository) GetWalletByID(id uuid.UUID) (*models.Wallet, error) {
	var wallet models.Wallet
	err := r.db.Where("id = ? AND is_active = ?", id, true).First(&wallet).Error
	return &wallet, err
}

func (r *walletRepository) GetWalletsByUserID(userID uuid.UUID) ([]*models.Wallet, error) {
	var wallets []*models.Wallet
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).Find(&wallets).Error
	return wallets, err
}

func (r *walletRepository) GetWalletsByClientUserID(clientUserID string) ([]*models.Wallet, error) {
	var wallets []*models.Wallet
	err := r.db.Where("client_user_id = ? AND is_active = ?", clientUserID, true).Find(&wallets).Error
	return wallets, err
}

func (r *walletRepository) GetWalletByAddress(address string) (*models.Wallet, error) {
	var wallet models.Wallet
	err := r.db.Where("address = ? AND is_active = ?", address, true).First(&wallet).Error
	return &wallet, err
}

func (r *walletRepository) GetWalletsByUserAndClient(userID uuid.UUID, clientUserID string) ([]*models.Wallet, error) {
	var wallets []*models.Wallet
	err := r.db.Where("user_id = ? AND client_user_id = ? AND is_active = ?", userID, clientUserID, true).Find(&wallets).Error
	return wallets, err
}

func (r *walletRepository) UpdateWallet(wallet *models.Wallet) error {
	return r.db.Save(wallet).Error
}

func (r *walletRepository) DeleteWallet(id uuid.UUID) error {
	return r.db.Delete(&models.Wallet{}, id).Error
}

func (r *walletRepository) DeactivateWallet(id uuid.UUID) error {
	return r.db.Model(&models.Wallet{}).Where("id = ?", id).Update("is_active", false).Error
}

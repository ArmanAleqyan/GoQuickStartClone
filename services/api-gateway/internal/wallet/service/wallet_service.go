package service

import (
	"fmt"

	"ironnode/pkg/crypto"
	"ironnode/pkg/models"
	"ironnode/services/api-gateway/internal/wallet/repository"

	"github.com/google/uuid"
)

type WalletService interface {
	CreateWallets(userID uuid.UUID, clientUserID string, purpose string, networks []string) ([]*models.WalletResponse, error)
	GetWalletsByUser(userID uuid.UUID, networks []string) ([]*models.WalletResponse, error)
	GetWalletsByClient(clientUserID string, networks []string) ([]*models.WalletResponse, error)
	GetWalletsByUserAndClient(userID uuid.UUID, clientUserID string, networks []string) ([]*models.WalletResponse, error)
	GetWalletByID(id uuid.UUID) (*models.WalletResponse, error)
	DeactivateWallet(id uuid.UUID, userID uuid.UUID) error
}

type walletService struct {
	repo              repository.WalletRepository
	encryptionService *crypto.EncryptionService
}

func NewWalletService(repo repository.WalletRepository, encryptionService *crypto.EncryptionService) WalletService {
	return &walletService{
		repo:              repo,
		encryptionService: encryptionService,
	}
}

// CreateWallets - создает кошельки для указанных сетей (или всех, если не указано)
func (s *walletService) CreateWallets(userID uuid.UUID, clientUserID string, purpose string, networks []string) ([]*models.WalletResponse, error) {
	// Если сети не указаны - создаем все
	if len(networks) == 0 {
		networks = []string{"ETH", "BTC", "BEP20", "TRC20", "MATIC"}
	}

	// Проверяем, какие кошельки уже существуют для этого клиента
	existingWallets, err := s.repo.GetWalletsByUserAndClient(userID, clientUserID)

	// Создаем map существующих сетей
	existingNetworks := make(map[models.NetworkType]bool)
	if err == nil {
		for _, wallet := range existingWallets {
			existingNetworks[wallet.Network] = true
		}
	}

	// Создаем только те кошельки, которых еще нет
	for _, network := range networks {
		networkType := models.NetworkType(network)

		// Если кошелек для этой сети уже существует - пропускаем
		if existingNetworks[networkType] {
			continue
		}

		var generateFunc func() (*crypto.WalletData, error)

		switch networkType {
		case models.NetworkETH:
			generateFunc = crypto.GenerateETHWallet
		case models.NetworkBTC:
			generateFunc = crypto.GenerateBTCWallet
		case models.NetworkBEP20:
			generateFunc = crypto.GenerateBEP20Wallet
		case models.NetworkTRC20:
			generateFunc = crypto.GenerateTRC20Wallet
		case models.NetworkMATIC:
			generateFunc = crypto.GenerateMATICWallet
		default:
			continue
		}

		_, err := s.createWallet(userID, clientUserID, purpose, networkType, generateFunc)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s wallet: %v", networkType, err)
		}
	}

	// Получаем все кошельки (существующие + только что созданные) и фильтруем по запрошенным сетям
	allWallets, err := s.repo.GetWalletsByUserAndClient(userID, clientUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallets: %v", err)
	}

	// Формируем массив ответов
	return s.buildArrayResponse(allWallets, networks), nil
}

// createWallet - helper для создания кошелька
func (s *walletService) createWallet(
	userID uuid.UUID,
	clientUserID string,
	purpose string,
	network models.NetworkType,
	generateFunc func() (*crypto.WalletData, error),
) (*models.Wallet, error) {
	// Генерируем кошелек
	walletData, err := generateFunc()
	if err != nil {
		return nil, err
	}

	// Шифруем приватный ключ
	encryptedKey, err := s.encryptionService.Encrypt(walletData.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %v", err)
	}

	// Создаем запись в БД
	wallet := &models.Wallet{
		UserID:              userID,
		ClientUserID:        clientUserID,
		Address:             walletData.Address,
		Network:             network,
		Purpose:             purpose,
		PublicKey:           walletData.PublicKey,
		HexAddress:          walletData.HexAddress,
		PrivateKeyEncrypted: encryptedKey,
		IsActive:            true,
	}

	if err := s.repo.CreateWallet(wallet); err != nil {
		return nil, err
	}

	return wallet, nil
}

// GetWalletsByUser - получить все кошельки пользователя
func (s *walletService) GetWalletsByUser(userID uuid.UUID, networks []string) ([]*models.WalletResponse, error) {
	wallets, err := s.repo.GetWalletsByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Создаем map для быстрой проверки нужных сетей
	networkMap := make(map[string]bool)
	for _, network := range networks {
		networkMap[network] = true
	}

	responses := []*models.WalletResponse{}
	for _, wallet := range wallets {
		// Если указаны конкретные сети - фильтруем
		if len(networks) > 0 && !networkMap[string(wallet.Network)] {
			continue
		}
		response := wallet.ToResponse()
		responses = append(responses, &response)
	}

	return responses, nil
}

// GetWalletsByClient - получить все кошельки клиента
func (s *walletService) GetWalletsByClient(clientUserID string, networks []string) ([]*models.WalletResponse, error) {
	wallets, err := s.repo.GetWalletsByClientUserID(clientUserID)
	if err != nil {
		return nil, err
	}

	// Создаем map для быстрой проверки нужных сетей
	networkMap := make(map[string]bool)
	for _, network := range networks {
		networkMap[network] = true
	}

	responses := []*models.WalletResponse{}
	for _, wallet := range wallets {
		// Если указаны конкретные сети - фильтруем
		if len(networks) > 0 && !networkMap[string(wallet.Network)] {
			continue
		}
		response := wallet.ToResponse()
		responses = append(responses, &response)
	}

	return responses, nil
}

// GetWalletsByUserAndClient - получить кошельки конкретного клиента конкретного пользователя
func (s *walletService) GetWalletsByUserAndClient(userID uuid.UUID, clientUserID string, networks []string) ([]*models.WalletResponse, error) {
	wallets, err := s.repo.GetWalletsByUserAndClient(userID, clientUserID)
	if err != nil {
		return nil, err
	}

	// Создаем map для быстрой проверки нужных сетей
	networkMap := make(map[string]bool)
	for _, network := range networks {
		networkMap[network] = true
	}

	responses := []*models.WalletResponse{}
	for _, wallet := range wallets {
		// Если указаны конкретные сети - фильтруем
		if len(networks) > 0 && !networkMap[string(wallet.Network)] {
			continue
		}
		response := wallet.ToResponse()
		responses = append(responses, &response)
	}

	return responses, nil
}

// GetWalletByID - получить кошелек по ID
func (s *walletService) GetWalletByID(id uuid.UUID) (*models.WalletResponse, error) {
	wallet, err := s.repo.GetWalletByID(id)
	if err != nil {
		return nil, err
	}

	response := wallet.ToResponse()
	return &response, nil
}

// DeactivateWallet - деактивировать кошелек (мягкое удаление)
func (s *walletService) DeactivateWallet(id uuid.UUID, userID uuid.UUID) error {
	// Проверяем что кошелек принадлежит пользователю
	wallet, err := s.repo.GetWalletByID(id)
	if err != nil {
		return fmt.Errorf("wallet not found: %v", err)
	}

	if wallet.UserID != userID {
		return fmt.Errorf("unauthorized: wallet does not belong to user")
	}

	return s.repo.DeactivateWallet(id)
}

// buildArrayResponse - формирует массив ответов из существующих кошельков
func (s *walletService) buildArrayResponse(wallets []*models.Wallet, networks []string) []*models.WalletResponse {
	// Создаем map для быстрой проверки нужных сетей
	networkMap := make(map[string]bool)
	for _, network := range networks {
		networkMap[network] = true
	}

	responses := []*models.WalletResponse{}
	for _, wallet := range wallets {
		// Если указаны конкретные сети - проверяем, входит ли текущий кошелек в список
		if len(networks) > 0 && !networkMap[string(wallet.Network)] {
			continue
		}

		walletResponse := wallet.ToResponse()
		responses = append(responses, &walletResponse)
	}

	return responses
}

package handler

import (
	"net/http"
	"strings"

	"ironnode/pkg/crypto"
	"ironnode/pkg/database"
	"ironnode/pkg/response"
	"ironnode/services/api-gateway/internal/wallet/repository"
	"ironnode/services/api-gateway/internal/wallet/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	walletService service.WalletService
)

func InitWalletService(dbDSN string) error {
	// Создаем подключение к БД
	db, err := database.NewPostgresConnection(dbDSN)
	if err != nil {
		return err
	}

	encryptionService, err := crypto.NewEncryptionService()
	if err != nil {
		return err
	}

	walletRepo := repository.NewWalletRepository(db)
	walletService = service.NewWalletService(walletRepo, encryptionService)
	return nil
}

// CreateWallet - создает кошельки для всех основных сетей (ETH, BTC, BEP20, TRC20, MATIC)
// POST /api/v1/wallets
func CreateWallet(c *gin.Context) {
	// Получаем user_id из JWT токена (middleware должен установить это)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.BadRequest(c, "Invalid user ID", err)
		return
	}

	// Получаем данные из запроса
	var req struct {
		ClientUserID string   `json:"user_id" binding:"required"` // ID клиента программиста
		Purpose      string   `json:"purpose"`                     // Назначение кошелька
		Networks     []string `json:"networks"`                    // Массив сетей (необязательно, если пусто - создаем все)
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err)
		return
	}

	// Создаем кошельки для указанных сетей (если не указано - для всех)
	wallets, err := walletService.CreateWallets(userID, req.ClientUserID, req.Purpose, req.Networks)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create wallets", err)
		return
	}

	// Возвращаем массив в единой структуре ответа
	response.Success(c, http.StatusCreated, "Wallets created successfully", wallets)
}

// GetWallets - получить все кошельки текущего пользователя
// GET /api/v1/wallets?networks=ETH,BTC (опционально)
func GetWallets(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.BadRequest(c, "Invalid user ID", err)
		return
	}

	// Получаем фильтр по сетям из query параметров
	networksParam := c.Query("networks")
	var networks []string
	if networksParam != "" {
		networks = strings.Split(networksParam, ",")
	}

	wallets, err := walletService.GetWalletsByUser(userID, networks)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get wallets", err)
		return
	}

	response.Success(c, http.StatusOK, "Wallets retrieved successfully", wallets)
}

// GetWalletsByClient - получить кошельки конкретного клиента
// GET /api/v1/wallets/client/:client_user_id?networks=ETH,BTC (опционально)
func GetWalletsByClient(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.BadRequest(c, "Invalid user ID", err)
		return
	}

	clientUserID := c.Param("client_user_id")
	if clientUserID == "" {
		response.BadRequest(c, "Client user ID is required", nil)
		return
	}

	// Получаем фильтр по сетям из query параметров
	networksParam := c.Query("networks")
	var networks []string
	if networksParam != "" {
		networks = strings.Split(networksParam, ",")
	}

	// Получаем кошельки конкретного клиента конкретного пользователя
	wallets, err := walletService.GetWalletsByUserAndClient(userID, clientUserID, networks)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get wallets", err)
		return
	}

	response.Success(c, http.StatusOK, "Client wallets retrieved successfully", wallets)
}

// GetWalletByID - получить кошелек по ID
// GET /api/v1/wallets/:id
func GetWalletByID(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	walletIDStr := c.Param("id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid wallet ID", err)
		return
	}

	wallet, err := walletService.GetWalletByID(walletID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Wallet not found", err)
		return
	}

	response.Success(c, http.StatusOK, "Wallet retrieved successfully", wallet)
}

// DeactivateWallet - деактивировать кошелек
// DELETE /api/v1/wallets/:id
func DeactivateWallet(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.BadRequest(c, "Invalid user ID", err)
		return
	}

	walletIDStr := c.Param("id")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid wallet ID", err)
		return
	}

	if err := walletService.DeactivateWallet(walletID, userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to deactivate wallet", err)
		return
	}

	response.Success(c, http.StatusOK, "Wallet deactivated successfully", map[string]interface{}{
		"id": walletID,
	})
}

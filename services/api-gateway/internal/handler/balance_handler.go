package handler

import (
	"net/http"

	"ironnode/pkg/response"
	"ironnode/pkg/tron"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	tronClient *tron.Client
)

// InitTronClient - инициализирует Tron клиент
func InitTronClient(nodeURL string) {
	tronClient = tron.NewClient(nodeURL)
}

// GetTronBalance - получить баланс TRX и USDT для адреса
// GET /api/v1/balance/tron/:address
func GetTronBalance(c *gin.Context) {
	address := c.Param("address")

	// Валидация адреса
	if !tron.ValidateTronAddress(address) {
		response.BadRequest(c, "Invalid Tron address. Address must start with 'T' and be 34 characters long", nil)
		return
	}

	// Получаем балансы
	balances, err := tronClient.GetBalances(address)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get balances", err)
		return
	}

	response.Success(c, http.StatusOK, "Balances retrieved successfully", balances)
}

// GetTRXBalanceOnly - получить только баланс TRX
// GET /api/v1/balance/trx/:address
func GetTRXBalanceOnly(c *gin.Context) {
	address := c.Param("address")

	// Валидация адреса
	if !tron.ValidateTronAddress(address) {
		response.BadRequest(c, "Invalid Tron address. Address must start with 'T' and be 34 characters long", nil)
		return
	}

	// Получаем TRX баланс
	trxBalance, err := tronClient.GetTRXBalance(address)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get TRX balance", err)
		return
	}

	// Конвертируем в TRX
	trxDecimal := tron.ConvertSunToTRX(trxBalance)

	result := gin.H{
		"address":     address,
		"trx_balance": trxDecimal,
	}

	response.Success(c, http.StatusOK, "TRX balance retrieved successfully", result)
}

// GetUSDTBalanceOnly - получить только баланс USDT TRC20
// GET /api/v1/balance/usdt/:address
func GetUSDTBalanceOnly(c *gin.Context) {
	address := c.Param("address")

	// Валидация адреса
	if !tron.ValidateTronAddress(address) {
		response.BadRequest(c, "Invalid Tron address. Address must start with 'T' and be 34 characters long", nil)
		return
	}

	// Получаем USDT баланс
	usdtBalance, err := tronClient.GetUSDTBalance(address)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get USDT balance", err)
		return
	}

	// Конвертируем в USDT
	usdtDecimal := tron.ConvertRawToUSDT(usdtBalance)

	result := gin.H{
		"address":      address,
		"usdt_balance": usdtDecimal,
	}

	response.Success(c, http.StatusOK, "USDT balance retrieved successfully", result)
}

// GetBalancesByAddress - получить балансы по адресу (POST версия)
// POST /api/v1/balance/check
// Body: {"address": "TYs..."}
func GetBalancesByAddress(c *gin.Context) {
	var req struct {
		Address string `json:"address" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request. 'address' field is required", err)
		return
	}

	// Валидация адреса
	if !tron.ValidateTronAddress(req.Address) {
		response.BadRequest(c, "Invalid Tron address. Address must start with 'T' and be 34 characters long", nil)
		return
	}

	// Получаем балансы
	balances, err := tronClient.GetBalances(req.Address)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get balances", err)
		return
	}

	response.Success(c, http.StatusOK, "Balances retrieved successfully", balances)
}

// GetBalanceByWalletID - получить баланс по ID кошелька из БД
// GET /api/v1/balance/wallet/:wallet_id
func GetBalanceByWalletID(c *gin.Context) {
	walletIDStr := c.Param("wallet_id")

	// Парсим UUID
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid wallet ID format", err)
		return
	}

	// Получаем кошелек из БД
	wallet, err := walletService.GetWalletByID(walletID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Wallet not found", err)
		return
	}

	// Проверяем что это TRC20 кошелек
	if wallet.Network != "TRC20" {
		response.BadRequest(c, "Only TRC20 wallets are supported for balance check", nil)
		return
	}

	// Валидация адреса
	if !tron.ValidateTronAddress(wallet.Address) {
		response.Error(c, http.StatusInternalServerError, "Invalid Tron address in database", nil)
		return
	}

	// Получаем балансы
	balances, err := tronClient.GetBalances(wallet.Address)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get balances", err)
		return
	}

	// Добавляем информацию о кошельке
	result := gin.H{
		"wallet_id":      wallet.ID,
		"client_user_id": wallet.ClientUserID,
		"address":        balances.Address,
		"network":        wallet.Network,
		"trx_balance":    balances.TRXBalance,
		"usdt_balance":   balances.USDTBalance,
	}

	response.Success(c, http.StatusOK, "Balances retrieved successfully", result)
}


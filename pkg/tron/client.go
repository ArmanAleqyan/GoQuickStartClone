package tron

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	tronaddress "github.com/fbsobreira/gotron-sdk/pkg/address"
)

const (
	// USDT TRC20 contract address
	USDTContractAddress = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
)

// Client - клиент для работы с Tron API
type Client struct {
	nodeURL    string
	httpClient *http.Client
}

// NewClient - создает новый Tron клиент
func NewClient(nodeURL string) *Client {
	return &Client{
		nodeURL: nodeURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// BalanceResponse - ответ с балансами
type BalanceResponse struct {
	Address     string `json:"address"`
	TRXBalance  string `json:"trx_balance"`  // В TRX (десятичный формат)
	USDTBalance string `json:"usdt_balance"` // В USDT (десятичный формат)
}

// GetBalances - получает балансы TRX и USDT для адреса
func (c *Client) GetBalances(address string) (*BalanceResponse, error) {
	// Получаем TRX баланс в SUN
	trxBalanceRaw, err := c.GetTRXBalance(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get TRX balance: %v", err)
	}

	// Получаем USDT баланс в минимальных единицах
	usdtBalanceRaw, err := c.GetUSDTBalance(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get USDT balance: %v", err)
	}

	// Конвертируем в десятичные значения
	trxDecimal := ConvertSunToTRX(trxBalanceRaw)
	usdtDecimal := ConvertRawToUSDT(usdtBalanceRaw)

	return &BalanceResponse{
		Address:     address,
		TRXBalance:  trxDecimal,
		USDTBalance: usdtDecimal,
	}, nil
}

// GetTRXBalance - получает баланс TRX
func (c *Client) GetTRXBalance(address string) (string, error) {
	// Конвертируем base58 адрес в hex
	hexAddress, err := c.base58ToHex(address)
	if err != nil {
		return "0", fmt.Errorf("invalid address: %v", err)
	}

	// Подготавливаем запрос
	requestBody := map[string]interface{}{
		"address": hexAddress,
		"visible": true, // Используем base58 адреса
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "0", fmt.Errorf("failed to marshal request: %v", err)
	}

	// Делаем запрос к ноде
	resp, err := c.httpClient.Post(
		c.nodeURL+"/wallet/getaccount",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "0", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "0", fmt.Errorf("failed to read response: %v", err)
	}

	// Парсим ответ
	var accountInfo map[string]interface{}
	if err := json.Unmarshal(body, &accountInfo); err != nil {
		return "0", fmt.Errorf("failed to parse response: %v", err)
	}

	// Проверяем наличие аккаунта
	if len(accountInfo) == 0 {
		return "0", nil // Аккаунт не существует - баланс 0
	}

	// Получаем баланс (в SUN)
	balance, ok := accountInfo["balance"].(float64)
	if !ok {
		return "0", nil // Баланс не найден - возвращаем 0
	}

	return fmt.Sprintf("%.0f", balance), nil
}

// GetUSDTBalance - получает баланс USDT TRC20
func (c *Client) GetUSDTBalance(address string) (string, error) {
	// Конвертируем base58 адрес в hex
	addr, err := tronaddress.Base58ToAddress(address)
	if err != nil {
		return "0", fmt.Errorf("failed to decode address: %v", err)
	}

	// Получаем hex адрес (с префиксом 41)
	hexAddress := hex.EncodeToString(addr.Bytes())

	// Конвертируем адрес контракта в hex
	contractAddr, err := tronaddress.Base58ToAddress(USDTContractAddress)
	if err != nil {
		return "0", fmt.Errorf("failed to decode contract address: %v", err)
	}
	hexContractAddress := hex.EncodeToString(contractAddr.Bytes())

	// Получаем параметр для balanceOf (адрес без префикса 41, дополненный до 64 символов)
	addressBytes := addr.Bytes()
	if len(addressBytes) > 0 && addressBytes[0] == 0x41 {
		addressBytes = addressBytes[1:]
	}
	parameter := padLeft(hex.EncodeToString(addressBytes), 64)

	// Подготавливаем запрос (БЕЗ visible, используем hex адреса)
	requestBody := map[string]interface{}{
		"owner_address":     hexAddress,
		"contract_address":  hexContractAddress,
		"function_selector": "balanceOf(address)",
		"parameter":         parameter,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "0", fmt.Errorf("failed to marshal request: %v", err)
	}

	// Делаем запрос к ноде (используем triggerconstantcontract для readonly операций)
	resp, err := c.httpClient.Post(
		c.nodeURL+"/wallet/triggerconstantcontract",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "0", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "0", fmt.Errorf("failed to read response: %v", err)
	}

	// Парсим ответ
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "0", fmt.Errorf("failed to parse response: %v", err)
	}

	// Проверяем результат
	constantResult, ok := result["constant_result"].([]interface{})
	if !ok || len(constantResult) == 0 {
		return "0", nil
	}

	// Получаем hex значение баланса
	balanceHex, ok := constantResult[0].(string)
	if !ok {
		return "0", nil
	}

	// Конвертируем hex в число
	balance := hexToDecimal(balanceHex)
	return balance, nil
}

// base58ToHex - конвертирует Tron base58 адрес в hex (упрощенная версия)
func (c *Client) base58ToHex(address string) (string, error) {
	// Для visible=true API Tron принимает base58 адреса напрямую
	// Просто возвращаем адрес как есть
	return address, nil
}

// addressToParameter - конвертирует адрес в параметр для smart contract (64 hex символа)
func (c *Client) addressToParameter(address string) (string, error) {
	// Используем gotron-sdk для конвертации base58 в Address
	addr, err := tronaddress.Base58ToAddress(address)
	if err != nil {
		return "", fmt.Errorf("failed to decode address: %v", err)
	}

	// Получаем байты адреса
	addressBytes := addr.Bytes()

	// Tron адреса в формате: [0x41][20 байт адреса]
	// Убираем префикс 0x41 (первый байт)
	if len(addressBytes) > 0 && addressBytes[0] == 0x41 {
		addressBytes = addressBytes[1:]
	}

	// Конвертируем в hex
	hexAddr := hex.EncodeToString(addressBytes)

	// Дополняем нулями слева до 64 символов
	return padLeft(hexAddr, 64), nil
}

// padLeft - дополняет строку нулями слева
func padLeft(str string, length int) string {
	if len(str) >= length {
		return str[len(str)-length:]
	}
	return strings.Repeat("0", length-len(str)) + str
}

// hexToDecimal - конвертирует hex строку в десятичное число
func hexToDecimal(hexStr string) string {
	// Убираем префикс 0x если есть
	hexStr = strings.TrimPrefix(hexStr, "0x")

	// Конвертируем в big.Int
	n := new(big.Int)
	n, ok := n.SetString(hexStr, 16)
	if !ok {
		return "0"
	}

	return n.String()
}

// ConvertSunToTRX - конвертирует SUN в TRX (1 TRX = 1,000,000 SUN)
func ConvertSunToTRX(sun string) string {
	n := new(big.Int)
	n, ok := n.SetString(sun, 10)
	if !ok {
		return "0"
	}

	// Делим на 1,000,000
	divisor := big.NewInt(1000000)
	result := new(big.Float).SetInt(n)
	divisorFloat := new(big.Float).SetInt(divisor)
	result.Quo(result, divisorFloat)

	return result.Text('f', 6)
}

// ConvertRawToUSDT - конвертирует raw единицы в USDT (1 USDT = 1,000,000 единиц)
func ConvertRawToUSDT(raw string) string {
	n := new(big.Int)
	n, ok := n.SetString(raw, 10)
	if !ok {
		return "0"
	}

	// Делим на 1,000,000
	divisor := big.NewInt(1000000)
	result := new(big.Float).SetInt(n)
	divisorFloat := new(big.Float).SetInt(divisor)
	result.Quo(result, divisorFloat)

	return result.Text('f', 6)
}

// ValidateTronAddress - проверяет валидность Tron адреса
func ValidateTronAddress(address string) bool {
	// Tron адреса начинаются с T и имеют длину 34 символа
	if !strings.HasPrefix(address, "T") {
		return false
	}

	if len(address) != 34 {
		return false
	}

	return true
}

// ConvertHexToBase58 - конвертирует hex адрес в base58 (для отображения)
func ConvertHexToBase58(hexAddr string) (string, error) {
	// Убираем 0x префикс
	hexAddr = strings.TrimPrefix(hexAddr, "0x")

	// Декодируем hex
	decoded, err := hex.DecodeString(hexAddr)
	if err != nil {
		return "", err
	}

	// В production здесь нужна полная реализация base58 кодирования
	// Для демо возвращаем упрощенную версию
	return "T" + hex.EncodeToString(decoded[:20]), nil
}

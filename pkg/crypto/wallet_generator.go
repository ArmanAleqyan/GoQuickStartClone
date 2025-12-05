package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
)

// WalletData - данные сгенерированного кошелька
type WalletData struct {
	Address    string // Публичный адрес
	PublicKey  string // Публичный ключ
	PrivateKey string // Приватный ключ (НЕ зашифрованный)
	HexAddress string // Адрес в hex формате
}

// GenerateETHWallet - генерирует новый Ethereum кошелек
func GenerateETHWallet() (*WalletData, error) {
	// Ethereum и BEP20 используют одну криптографию
	return generateEVMWallet()
}

// GenerateBEP20Wallet - генерирует новый BEP20 кошелек (Binance Smart Chain)
// BEP20 использует ту же криптографию что и Ethereum (secp256k1)
func GenerateBEP20Wallet() (*WalletData, error) {
	return generateEVMWallet()
}

// GenerateMATICWallet - генерирует новый Polygon (MATIC) кошелек
func GenerateMATICWallet() (*WalletData, error) {
	// Polygon использует ту же криптографию что и Ethereum
	return generateEVMWallet()
}

// generateEVMWallet - общий генератор для EVM-совместимых сетей (ETH, BSC, Polygon)
func generateEVMWallet() (*WalletData, error) {
	// Генерируем приватный ключ
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	// Получаем приватный ключ в hex формате
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	// Получаем публичный ключ
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key to ECDSA")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	publicKeyHex := hex.EncodeToString(publicKeyBytes)

	// Получаем адрес
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	addressHex := address.Hex()

	return &WalletData{
		Address:    addressHex,
		PublicKey:  publicKeyHex,
		PrivateKey: privateKeyHex,
		HexAddress: addressHex,
	}, nil
}

// GenerateBTCWallet - генерирует новый Bitcoin кошелек
func GenerateBTCWallet() (*WalletData, error) {
	// Генерируем приватный ключ (Bitcoin использует ту же secp256k1 кривую)
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate BTC private key: %v", err)
	}

	// Получаем приватный ключ в hex формате
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	// Получаем публичный ключ
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key to ECDSA")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	publicKeyHex := hex.EncodeToString(publicKeyBytes)

	// Для BTC адреса нужно использовать специальное форматирование
	// Здесь упрощенная версия - в production нужна btcd библиотека
	// Для демо просто используем hash160 от публичного ключа
	address := generateBTCAddress(publicKeyBytes)

	return &WalletData{
		Address:    address,
		PublicKey:  publicKeyHex,
		PrivateKey: privateKeyHex,
		HexAddress: address,
	}, nil
}

// generateBTCAddress - генерирует Bitcoin адрес из публичного ключа
// Упрощенная версия для демо
func generateBTCAddress(publicKeyBytes []byte) string {
	// В production здесь должна быть полная реализация с base58check
	// Для демо возвращаем префикс + hash
	hash := crypto.Keccak256(publicKeyBytes)
	return "1" + hex.EncodeToString(hash[:20]) // Prefix "1" для mainnet
}

// GenerateTRC20Wallet - генерирует новый TRC20 кошелек (Tron)
func GenerateTRC20Wallet() (*WalletData, error) {
	// Генерируем приватный ключ для Tron (используем ту же криптографию что и Ethereum)
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Tron private key: %v", err)
	}

	// Получаем приватный ключ в hex формате
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	// Получаем публичный ключ
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key to ECDSA")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	publicKeyHex := hex.EncodeToString(publicKeyBytes)

	// Получаем Tron адрес (base58)
	tronAddress := address.PubkeyToAddress(*publicKeyECDSA)
	addressStr := tronAddress.String()

	// Получаем hex адрес
	hexAddress := hex.EncodeToString(tronAddress.Bytes())

	return &WalletData{
		Address:    addressStr,      // Base58 адрес (T...)
		PublicKey:  publicKeyHex,
		PrivateKey: privateKeyHex,
		HexAddress: "0x" + hexAddress, // Hex формат
	}, nil
}

// ValidatePrivateKey - проверяет валидность приватного ключа
func ValidatePrivateKey(privateKeyHex string) error {
	// Декодируем hex
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return fmt.Errorf("invalid hex format: %v", err)
	}

	// Проверяем длину (должно быть 32 байта)
	if len(privateKeyBytes) != 32 {
		return fmt.Errorf("invalid private key length: expected 32 bytes, got %d", len(privateKeyBytes))
	}

	return nil
}

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

// EncryptionService - сервис для шифрования/дешифрования приватных ключей
type EncryptionService struct {
	encryptionKey []byte
}

// NewEncryptionService - создает новый сервис шифрования
// Ключ шифрования должен быть 32 байта для AES-256
func NewEncryptionService() (*EncryptionService, error) {
	// Получаем ключ шифрования из переменной окружения
	keyString := os.Getenv("ENCRYPTION_KEY")
	if keyString == "" {
		return nil, fmt.Errorf("ENCRYPTION_KEY environment variable is not set")
	}

	// Декодируем ключ из base64
	key, err := base64.StdEncoding.DecodeString(keyString)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption key: %v", err)
	}

	// Проверяем длину ключа (должен быть 32 байта для AES-256)
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes for AES-256, got %d bytes", len(key))
	}

	return &EncryptionService{
		encryptionKey: key,
	}, nil
}

// Encrypt - шифрует данные используя AES-256-GCM
func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	// Создаем GCM режим
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	// Создаем nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %v", err)
	}

	// Шифруем данные
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Возвращаем в base64 для удобного хранения в БД
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt - расшифровывает данные
func (s *EncryptionService) Decrypt(ciphertext string) (string, error) {
	// Декодируем из base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %v", err)
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// Извлекаем nonce и ciphertext
	nonce, encryptedData := data[:nonceSize], data[nonceSize:]

	// Расшифровываем
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %v", err)
	}

	return string(plaintext), nil
}

// GenerateEncryptionKey - генерирует новый ключ шифрования (для первоначальной настройки)
// Этот метод нужно вызвать один раз для генерации ключа и сохранить его в переменную окружения
func GenerateEncryptionKey() (string, error) {
	key := make([]byte, 32) // 32 байта = 256 бит для AES-256
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate key: %v", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

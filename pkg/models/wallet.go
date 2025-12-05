package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NetworkType string

const (
	NetworkETH    NetworkType = "ETH"     // Ethereum Mainnet
	NetworkBTC    NetworkType = "BTC"     // Bitcoin Mainnet
	NetworkBEP20  NetworkType = "BEP20"   // Binance Smart Chain
	NetworkTRC20  NetworkType = "TRC20"   // Tron Network
	NetworkMATIC  NetworkType = "MATIC"   // Polygon Network
	NetworkSOL    NetworkType = "SOL"     // Solana Network
)

type Wallet struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID              uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`                  // Наш пользователь (владелец API ключа, из токена)
	ClientUserID        string         `gorm:"not null;index" json:"client_user_id"`                     // ID клиента из запроса (для кого создается кошелек)
	Address             string         `gorm:"uniqueIndex;not null" json:"address"`                      // Публичный адрес кошелька
	Network             NetworkType    `gorm:"type:varchar(10);not null;index" json:"network"`           // BEP20 или TRC20
	Purpose             string         `gorm:"type:varchar(255)" json:"purpose"`                         // Назначение кошелька
	PublicKey           string         `gorm:"type:text;not null" json:"public_key"`                     // Публичный ключ
	HexAddress          string         `gorm:"type:varchar(255)" json:"hex_address"`                     // Адрес в hex формате
	PrivateKeyEncrypted string         `gorm:"type:text;not null" json:"-"`                              // Зашифрованный приватный ключ (НЕ отдаем в JSON)
	IsActive            bool           `gorm:"default:true" json:"is_active"`                            // Активен ли кошелек
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
	User                User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// WalletResponse - структура для ответа API (без приватных данных)
type WalletResponse struct {
	ID           uuid.UUID   `json:"id"`
	ClientUserID string      `json:"client_user_id"`
	Address      string      `json:"address"`
	Network      NetworkType `json:"network"`
	CreatedAt    time.Time   `json:"created_at"`
}

func (w *Wallet) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

// ToResponse - конвертирует Wallet в WalletResponse (без приватных данных)
func (w *Wallet) ToResponse() WalletResponse {
	return WalletResponse{
		ID:           w.ID,
		ClientUserID: w.ClientUserID,
		Address:      w.Address,
		Network:      w.Network,
		CreatedAt:    w.CreatedAt,
	}
}

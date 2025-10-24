package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BlockchainType string

const (
	Ethereum BlockchainType = "ethereum"
	Bitcoin  BlockchainType = "bitcoin"
	Polygon  BlockchainType = "polygon"
	BSC      BlockchainType = "bsc"
	Avalanche BlockchainType = "avalanche"
	Solana   BlockchainType = "solana"
)

type BlockchainNode struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `json:"name"`
	Type        BlockchainType `gorm:"type:varchar(50);not null" json:"type"`
	Network     string         `json:"network"` // mainnet, testnet, etc.
	URL         string         `json:"url"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	Priority    int            `gorm:"default:0" json:"priority"` // Higher priority nodes are used first
	MaxRequests int            `gorm:"default:1000" json:"max_requests"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (b *BlockchainNode) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

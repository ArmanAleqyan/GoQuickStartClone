package repository

import (
	"ironnode/pkg/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NodeRepository interface {
	CreateNode(node *models.BlockchainNode) error
	GetNodeByID(id uuid.UUID) (*models.BlockchainNode, error)
	GetNodesByType(blockchainType models.BlockchainType) ([]*models.BlockchainNode, error)
	GetActiveNodes() ([]*models.BlockchainNode, error)
	UpdateNode(node *models.BlockchainNode) error
	DeleteNode(id uuid.UUID) error
}

type nodeRepository struct {
	db *gorm.DB
}

func NewNodeRepository(db *gorm.DB) NodeRepository {
	return &nodeRepository{db: db}
}

func (r *nodeRepository) CreateNode(node *models.BlockchainNode) error {
	return r.db.Create(node).Error
}

func (r *nodeRepository) GetNodeByID(id uuid.UUID) (*models.BlockchainNode, error) {
	var node models.BlockchainNode
	err := r.db.Where("id = ?", id).First(&node).Error
	return &node, err
}

func (r *nodeRepository) GetNodesByType(blockchainType models.BlockchainType) ([]*models.BlockchainNode, error) {
	var nodes []*models.BlockchainNode
	err := r.db.Where("type = ? AND is_active = ?", blockchainType, true).
		Order("priority DESC").
		Find(&nodes).Error
	return nodes, err
}

func (r *nodeRepository) GetActiveNodes() ([]*models.BlockchainNode, error) {
	var nodes []*models.BlockchainNode
	err := r.db.Where("is_active = ?", true).Order("priority DESC").Find(&nodes).Error
	return nodes, err
}

func (r *nodeRepository) UpdateNode(node *models.BlockchainNode) error {
	return r.db.Save(node).Error
}

func (r *nodeRepository) DeleteNode(id uuid.UUID) error {
	return r.db.Delete(&models.BlockchainNode{}, id).Error
}

package service

import (
	"ironnode/pkg/models"
	"ironnode/services/blockchain-service/internal/repository"

	"github.com/google/uuid"
)

type NodeService interface {
	CreateNode(name string, blockchainType models.BlockchainType, network, url string) (*models.BlockchainNode, error)
	GetNodeByID(id uuid.UUID) (*models.BlockchainNode, error)
	GetNodesByType(blockchainType models.BlockchainType) ([]*models.BlockchainNode, error)
	GetActiveNodes() ([]*models.BlockchainNode, error)
	UpdateNode(node *models.BlockchainNode) error
	DeleteNode(id uuid.UUID) error
}

type nodeService struct {
	repo repository.NodeRepository
}

func NewNodeService(repo repository.NodeRepository) NodeService {
	return &nodeService{repo: repo}
}

func (s *nodeService) CreateNode(name string, blockchainType models.BlockchainType, network, url string) (*models.BlockchainNode, error) {
	node := &models.BlockchainNode{
		Name:     name,
		Type:     blockchainType,
		Network:  network,
		URL:      url,
		IsActive: true,
		Priority: 0,
	}

	if err := s.repo.CreateNode(node); err != nil {
		return nil, err
	}

	return node, nil
}

func (s *nodeService) GetNodeByID(id uuid.UUID) (*models.BlockchainNode, error) {
	return s.repo.GetNodeByID(id)
}

func (s *nodeService) GetNodesByType(blockchainType models.BlockchainType) ([]*models.BlockchainNode, error) {
	return s.repo.GetNodesByType(blockchainType)
}

func (s *nodeService) GetActiveNodes() ([]*models.BlockchainNode, error) {
	return s.repo.GetActiveNodes()
}

func (s *nodeService) UpdateNode(node *models.BlockchainNode) error {
	return s.repo.UpdateNode(node)
}

func (s *nodeService) DeleteNode(id uuid.UUID) error {
	return s.repo.DeleteNode(id)
}

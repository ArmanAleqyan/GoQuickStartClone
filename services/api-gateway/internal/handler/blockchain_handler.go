package handler

import (
	"net/http"

	"ironnode/pkg/config"
	"ironnode/pkg/response"

	"github.com/gin-gonic/gin"
)

type BlockchainHandler struct {
	config *config.Config
}

func NewBlockchainHandler(cfg *config.Config) *BlockchainHandler {
	return &BlockchainHandler{
		config: cfg,
	}
}

func (h *BlockchainHandler) ListNodes(c *gin.Context) {
	// This would typically call the Blockchain Service via gRPC
	// For now, returning mock data
	nodes := []gin.H{
		{
			"id":       "1",
			"name":     "Ethereum Mainnet",
			"type":     "ethereum",
			"network":  "mainnet",
			"is_active": true,
		},
		{
			"id":       "2",
			"name":     "Polygon Mainnet",
			"type":     "polygon",
			"network":  "mainnet",
			"is_active": true,
		},
	}

	response.Success(c, http.StatusOK, "Nodes retrieved successfully", nodes)
}

func (h *BlockchainHandler) GetNode(c *gin.Context) {
	nodeID := c.Param("id")

	// This would typically call the Blockchain Service via gRPC
	node := gin.H{
		"id":       nodeID,
		"name":     "Ethereum Mainnet",
		"type":     "ethereum",
		"network":  "mainnet",
		"is_active": true,
	}

	response.Success(c, http.StatusOK, "Node retrieved successfully", node)
}

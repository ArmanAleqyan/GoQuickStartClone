package handler

import (
	"net/http"

	"ironnode/pkg/response"

	"github.com/gin-gonic/gin"
)

func ListAPIKeys(c *gin.Context) {
	// This would call User Service
	keys := []gin.H{
		{
			"id":          "1",
			"name":        "Production API Key",
			"key":         "qn_***************",
			"is_active":   true,
			"created_at":  "2024-01-01T00:00:00Z",
		},
		{
			"id":          "2",
			"name":        "Development API Key",
			"key":         "qn_***************",
			"is_active":   true,
			"created_at":  "2024-01-10T00:00:00Z",
		},
	}

	response.Success(c, http.StatusOK, "API keys retrieved", keys)
}

func CreateAPIKey(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err)
		return
	}

	// This would call User Service to create API key
	newKey := gin.H{
		"id":          "3",
		"name":        req.Name,
		"description": req.Description,
		"key":         "qn_abc123def456ghi789jkl",
		"is_active":   true,
		"created_at":  "2024-01-15T00:00:00Z",
	}

	response.Success(c, http.StatusCreated, "API key created", newKey)
}

func DeleteAPIKey(c *gin.Context) {
	keyID := c.Param("id")

	// This would call User Service to delete API key
	response.Success(c, http.StatusOK, "API key deleted", gin.H{
		"id": keyID,
	})
}

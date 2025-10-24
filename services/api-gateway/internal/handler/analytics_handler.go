package handler

import (
	"net/http"

	"ironnode/pkg/response"

	"github.com/gin-gonic/gin"
)

func GetUsageStats(c *gin.Context) {
	// This would call Analytics Service
	stats := gin.H{
		"total_requests":    1500,
		"requests_today":    150,
		"requests_this_month": 4500,
		"success_rate":      98.5,
		"average_response_time": 120, // ms
	}

	response.Success(c, http.StatusOK, "Usage stats retrieved", stats)
}

func GetRequestHistory(c *gin.Context) {
	// This would call Analytics Service
	history := []gin.H{
		{
			"id":            "1",
			"blockchain":    "ethereum",
			"method":        "eth_getBlockByNumber",
			"status_code":   200,
			"response_time": 125,
			"timestamp":     "2024-01-15T10:30:00Z",
		},
		{
			"id":            "2",
			"blockchain":    "polygon",
			"method":        "eth_call",
			"status_code":   200,
			"response_time": 98,
			"timestamp":     "2024-01-15T10:29:00Z",
		},
	}

	response.Success(c, http.StatusOK, "Request history retrieved", history)
}

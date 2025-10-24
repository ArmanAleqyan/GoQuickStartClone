package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RequestLog struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	APIKeyID      uuid.UUID `gorm:"type:uuid;index" json:"api_key_id"`
	Blockchain    string    `gorm:"index" json:"blockchain"`
	Method        string    `json:"method"`
	Endpoint      string    `json:"endpoint"`
	StatusCode    int       `json:"status_code"`
	ResponseTime  int64     `json:"response_time"` // in milliseconds
	RequestSize   int64     `json:"request_size"`  // in bytes
	ResponseSize  int64     `json:"response_size"` // in bytes
	IPAddress     string    `json:"ip_address"`
	UserAgent     string    `json:"user_agent"`
	Error         string    `json:"error,omitempty"`
	CreatedAt     time.Time `gorm:"index" json:"created_at"`
}

func (r *RequestLog) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

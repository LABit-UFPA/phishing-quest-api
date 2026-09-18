package dto

import "github.com/google/uuid"

// ConsentRequestDTO e o corpo de POST /api/v1/auth/consent.
type ConsentRequestDTO struct {
	UserID uuid.UUID `json:"userId" binding:"required"`
}

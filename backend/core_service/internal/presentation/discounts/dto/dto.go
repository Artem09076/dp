package dto

import "time"

type CreateDiscountRequest struct {
	Type      string    `json:"type"`
	Value     int       `json:"value"`
	ValidFrom time.Time `json:"valid_from"`
	ValidTo   time.Time `json:"valid_to"`
	MaxUses   int       `json:"max_uses"`
}

type PatchDiscountRequest struct {
	ValidFrom *time.Time `json:"valid_from,omitempty"`
	ValidTo   *time.Time `json:"valid_to,omitempty"`
	MaxUses   *int       `json:"max_uses,omitempty"`
}

type DiscountResponse struct {
	ID        string    `json:"id"`
	ServiceID string    `json:"service_id"`
	Type      string    `json:"type"`
	Value     int       `json:"value"`
	ValidFrom time.Time `json:"valid_from"`
	ValidTo   time.Time `json:"valid_to"`
	MaxUses   int       `json:"max_uses"`
	UsedCount int       `json:"used_count"`
	CreatedAt time.Time `json:"created_at"`
}

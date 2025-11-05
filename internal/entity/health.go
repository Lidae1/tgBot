package entity

import "time"

type HealStatus string

const (
	HealStatusOk       HealStatus = "ok"
	HealStatusError    HealStatus = "error"
	HealStatusDegraded HealStatus = "degraded"
)

type HealthResponse struct {
	Status    HealStatus           `json:"status"`
	Timestamp time.Time            `json:"timestamp"`
	Checks    map[string]HealCheck `json:"checks"`
}

type HealCheck struct {
	Status    HealStatus `json:"status"`
	Message   string     `json:"message"`
	Timestamp time.Time  `json:"timestamp"`
}

func NewHealthResponse() *HealthResponse {
	return &HealthResponse{
		Status:    HealStatusOk,
		Timestamp: time.Now(),
		Checks:    make(map[string]HealCheck),
	}
}

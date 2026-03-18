package models

import "time"

type Tweet struct {
	ID        uint64    `json:"id"`
	UserID    uint64    `json:"user_id"`
	Tweet     string    `json:"tweet"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

package models

import "time"

type Tweet struct {
	Id        int64     `json:"id"`
	UserId    int64     `json:"user_id"`
	Tweet     string    `json:"tweet"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Tags      []*Tag    `json:"tags,omitempty"`
}

package models

import "time"

type MediaAttachment struct {
	Id        int64     `json:"id"`
	UserId    int64     `json:"user_id"`
	TweetId   *int64    `json:"tweet_id,omitempty"`
	S3Key     string    `json:"s3_key"`
	Url       string    `json:"url"`
	MimeType  string    `json:"mime_type"`
	SizeBytes int64     `json:"size_bytes"`
	CreatedAt time.Time `json:"created_at"`
}

type MediaUpload struct {
	Attachment *MediaAttachment  `json:"attachment"`
	UploadURL  string            `json:"upload_url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers,omitempty"`
	ExpiresIn  int64             `json:"expires_in_seconds"`
}

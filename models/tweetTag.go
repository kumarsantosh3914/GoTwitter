package models

type TweetTag struct {
	TweetID uint64 `json:"tweet_id"`
	TagID   uint64 `json:"tag_id"`
}

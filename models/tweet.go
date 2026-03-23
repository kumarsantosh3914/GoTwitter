package models

import "time"

type Tweet struct {
	Id            int64              `json:"id"`
	UserId        int64              `json:"user_id"`
	ParentTweetID *int64             `json:"parent_tweet_id,omitempty"`
	Tweet         string             `json:"tweet"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	LikeCount     int64              `json:"like_count"`
	RetweetCount  int64              `json:"retweet_count"`
	ReplyCount    int64              `json:"reply_count"`
	IsLiked       bool               `json:"is_liked"`
	IsRetweeted   bool               `json:"is_retweeted"`
	Tags          []*Tag             `json:"tags,omitempty"`
	Media         []*MediaAttachment `json:"media,omitempty"`
	Replies       []*Tweet           `json:"replies,omitempty"`
	Thread        []*Tweet           `json:"thread,omitempty"`
}

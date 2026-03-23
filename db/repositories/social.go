package db

import (
	"context"
	"fmt"
	"strings"
)

type SocialRepository interface {
	LikeTweet(ctx context.Context, tweetID int64, userID int64) error
	UnlikeTweet(ctx context.Context, tweetID int64, userID int64) error
	RetweetTweet(ctx context.Context, tweetID int64, userID int64) error
	UnretweetTweet(ctx context.Context, tweetID int64, userID int64) error
	GetTweetInteractionStates(ctx context.Context, userID int64, tweetIDs []int64) (map[int64]TweetInteractionState, error)
	FollowUser(ctx context.Context, followerID int64, followeeID int64) error
	UnfollowUser(ctx context.Context, followerID int64, followeeID int64) error
	IsFollowing(ctx context.Context, followerID int64, followeeID int64) (bool, error)
	ListFollowers(ctx context.Context, userID int64, limit int, offset int) ([]int64, error)
	ListFollowing(ctx context.Context, userID int64, limit int, offset int) ([]int64, error)
	GetUserFollowStates(ctx context.Context, followerID int64, userIDs []int64) (map[int64]bool, error)
}

type TweetInteractionState struct {
	IsLiked     bool
	IsRetweeted bool
}

type SocialRepositoryImpl struct {
	db queryExecutor
}

func NewSocialRepository(db queryExecutor) SocialRepository {
	return &SocialRepositoryImpl{db: db}
}

func (s *SocialRepositoryImpl) LikeTweet(ctx context.Context, tweetID int64, userID int64) error {
	_, err := s.db.ExecContext(ctx, `INSERT IGNORE INTO tweet_likes (tweet_id, user_id) VALUES (?, ?)`, tweetID, userID)
	return err
}

func (s *SocialRepositoryImpl) UnlikeTweet(ctx context.Context, tweetID int64, userID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM tweet_likes WHERE tweet_id = ? AND user_id = ?`, tweetID, userID)
	return err
}

func (s *SocialRepositoryImpl) RetweetTweet(ctx context.Context, tweetID int64, userID int64) error {
	_, err := s.db.ExecContext(ctx, `INSERT IGNORE INTO tweet_retweets (tweet_id, user_id) VALUES (?, ?)`, tweetID, userID)
	return err
}

func (s *SocialRepositoryImpl) UnretweetTweet(ctx context.Context, tweetID int64, userID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM tweet_retweets WHERE tweet_id = ? AND user_id = ?`, tweetID, userID)
	return err
}

func (s *SocialRepositoryImpl) GetTweetInteractionStates(ctx context.Context, userID int64, tweetIDs []int64) (map[int64]TweetInteractionState, error) {
	if len(tweetIDs) == 0 {
		return map[int64]TweetInteractionState{}, nil
	}

	states := make(map[int64]TweetInteractionState, len(tweetIDs))
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(tweetIDs)), ",")

	args := make([]interface{}, 0, len(tweetIDs)+1)
	args = append(args, userID)
	for _, tweetID := range tweetIDs {
		args = append(args, tweetID)
	}

	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(
		`SELECT tweet_id FROM tweet_likes WHERE user_id = ? AND tweet_id IN (%s)`,
		placeholders,
	), args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var tweetID int64
		if err := rows.Scan(&tweetID); err != nil {
			rows.Close()
			return nil, err
		}
		state := states[tweetID]
		state.IsLiked = true
		states[tweetID] = state
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	rows, err = s.db.QueryContext(ctx, fmt.Sprintf(
		`SELECT tweet_id FROM tweet_retweets WHERE user_id = ? AND tweet_id IN (%s)`,
		placeholders,
	), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tweetID int64
		if err := rows.Scan(&tweetID); err != nil {
			return nil, err
		}
		state := states[tweetID]
		state.IsRetweeted = true
		states[tweetID] = state
	}
	return states, rows.Err()
}

func (s *SocialRepositoryImpl) FollowUser(ctx context.Context, followerID int64, followeeID int64) error {
	_, err := s.db.ExecContext(ctx, `INSERT IGNORE INTO user_follows (follower_id, followee_id) VALUES (?, ?)`, followerID, followeeID)
	return err
}

func (s *SocialRepositoryImpl) UnfollowUser(ctx context.Context, followerID int64, followeeID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM user_follows WHERE follower_id = ? AND followee_id = ?`, followerID, followeeID)
	return err
}

func (s *SocialRepositoryImpl) IsFollowing(ctx context.Context, followerID int64, followeeID int64) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM user_follows WHERE follower_id = ? AND followee_id = ?)`,
		followerID,
		followeeID,
	).Scan(&exists)
	return exists, err
}

func (s *SocialRepositoryImpl) ListFollowers(ctx context.Context, userID int64, limit int, offset int) ([]int64, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT follower_id FROM user_follows WHERE followee_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *SocialRepositoryImpl) ListFollowing(ctx context.Context, userID int64, limit int, offset int) ([]int64, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT followee_id FROM user_follows WHERE follower_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *SocialRepositoryImpl) GetUserFollowStates(ctx context.Context, followerID int64, userIDs []int64) (map[int64]bool, error) {
	if len(userIDs) == 0 {
		return map[int64]bool{}, nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(userIDs)), ",")
	args := make([]interface{}, 0, len(userIDs)+1)
	args = append(args, followerID)
	for _, id := range userIDs {
		args = append(args, id)
	}

	rows, err := s.db.QueryContext(
		ctx,
		fmt.Sprintf(`SELECT followee_id FROM user_follows WHERE follower_id = ? AND followee_id IN (%s)`, placeholders),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states := make(map[int64]bool, len(userIDs))
	for rows.Next() {
		var followeeID int64
		if err := rows.Scan(&followeeID); err != nil {
			return nil, err
		}
		states[followeeID] = true
	}
	return states, rows.Err()
}

package db

import (
	"GoTwitter/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type queryExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type TweetRepository interface {
	Create(ctx context.Context, tweet *models.Tweet) (*models.Tweet, error)
	GetByID(ctx context.Context, id int64) (*models.Tweet, error)
	GetAll(ctx context.Context, limit, offset int, userId int64, tag string, search string) ([]*models.Tweet, error)
	GetRepliesByParentIDs(ctx context.Context, parentIDs []int64) (map[int64][]*models.Tweet, error)
	Update(ctx context.Context, tweet *models.Tweet) error
	DeleteByID(ctx context.Context, id int64) error
}

type TweetRepositoryImpl struct {
	db queryExecutor
}

func NewTweetRepository(_db queryExecutor) TweetRepository {
	return &TweetRepositoryImpl{
		db: _db,
	}
}

func (t *TweetRepositoryImpl) Create(ctx context.Context, tweet *models.Tweet) (*models.Tweet, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if tweet == nil {
		return nil, errors.New("tweet is nil")
	}
	if t.db == nil {
		return nil, errors.New("db is nil")
	}

	res, err := t.db.ExecContext(
		ctx,
		`INSERT INTO tweets (user_id, parent_tweet_id, tweet, created_at, updated_at)
		 VALUES (?, ?, ?, NOW(), NOW())`,
		tweet.UserId,
		tweet.ParentTweetID,
		tweet.Tweet,
	)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	created, err := t.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (t *TweetRepositoryImpl) GetByID(ctx context.Context, id int64) (*models.Tweet, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if t.db == nil {
		return nil, errors.New("db is nil")
	}

	var tweet models.Tweet
	var parentTweetID sql.NullInt64
	err := t.db.QueryRowContext(ctx, tweetSelectQuery(`WHERE t.id = ? LIMIT 1`), id).Scan(
		&tweet.Id,
		&tweet.UserId,
		&parentTweetID,
		&tweet.Tweet,
		&tweet.CreatedAt,
		&tweet.UpdatedAt,
		&tweet.LikeCount,
		&tweet.RetweetCount,
		&tweet.ReplyCount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if parentTweetID.Valid {
		tweet.ParentTweetID = &parentTweetID.Int64
	}
	return &tweet, nil
}

func (t *TweetRepositoryImpl) GetAll(ctx context.Context, limit, offset int, userId int64, tag string, search string) ([]*models.Tweet, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if t.db == nil {
		return nil, errors.New("db is nil")
	}

	query := tweetSelectQuery("")

	var conditions []string
	var args []interface{}

	conditions = append(conditions, "t.parent_tweet_id IS NULL")

	if tag != "" {
		query += ` JOIN tweet_tags tt ON t.id = tt.tweet_id
		           JOIN tags tg ON tt.tag_id = tg.id`
		conditions = append(conditions, "tg.name = ?")
		args = append(args, tag)
	}

	if userId > 0 {
		conditions = append(conditions, "t.user_id = ?")
		args = append(args, userId)
	}

	if search != "" {
		conditions = append(conditions, "t.tweet LIKE ?")
		args = append(args, "%"+search+"%")
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY t.id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := t.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tweets []*models.Tweet
	for rows.Next() {
		tweet, err := scanTweet(rows)
		if err != nil {
			return nil, err
		}
		tweets = append(tweets, tweet)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tweets, nil
}

func (t *TweetRepositoryImpl) GetRepliesByParentIDs(ctx context.Context, parentIDs []int64) (map[int64][]*models.Tweet, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if t.db == nil {
		return nil, errors.New("db is nil")
	}
	if len(parentIDs) == 0 {
		return map[int64][]*models.Tweet{}, nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(parentIDs)), ",")
	args := make([]interface{}, 0, len(parentIDs))
	for _, id := range parentIDs {
		args = append(args, id)
	}

	rows, err := t.db.QueryContext(
		ctx,
		fmt.Sprintf("%s WHERE t.parent_tweet_id IN (%s) ORDER BY t.created_at ASC", tweetSelectQuery(""), placeholders),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	repliesByParent := make(map[int64][]*models.Tweet, len(parentIDs))
	for rows.Next() {
		tweet, err := scanTweet(rows)
		if err != nil {
			return nil, err
		}
		if tweet.ParentTweetID == nil {
			continue
		}
		repliesByParent[*tweet.ParentTweetID] = append(repliesByParent[*tweet.ParentTweetID], tweet)
	}

	return repliesByParent, rows.Err()
}

func (t *TweetRepositoryImpl) Update(ctx context.Context, tweet *models.Tweet) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if tweet == nil {
		return errors.New("tweet is nil")
	}
	if t.db == nil {
		return errors.New("db is nil")
	}

	_, err := t.db.ExecContext(
		ctx,
		`UPDATE tweets 
		 SET tweet = ?, updated_at = NOW()
		 WHERE id = ?`,
		tweet.Tweet,
		tweet.Id,
	)
	return err
}

func (t *TweetRepositoryImpl) DeleteByID(ctx context.Context, id int64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if t.db == nil {
		return errors.New("db is nil")
	}
	_, err := t.db.ExecContext(ctx, `DELETE FROM tweets WHERE id = ?`, id)
	return err
}

func tweetSelectQuery(suffix string) string {
	return `SELECT DISTINCT
		t.id,
		t.user_id,
		t.parent_tweet_id,
		t.tweet,
		t.created_at,
		t.updated_at,
		(SELECT COUNT(*) FROM tweet_likes tl WHERE tl.tweet_id = t.id) AS like_count,
		(SELECT COUNT(*) FROM tweet_retweets tr WHERE tr.tweet_id = t.id) AS retweet_count,
		(SELECT COUNT(*) FROM tweets replies WHERE replies.parent_tweet_id = t.id) AS reply_count
	FROM tweets t ` + suffix
}

func scanTweet(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.Tweet, error) {
	var tweet models.Tweet
	var parentTweetID sql.NullInt64

	if err := scanner.Scan(
		&tweet.Id,
		&tweet.UserId,
		&parentTweetID,
		&tweet.Tweet,
		&tweet.CreatedAt,
		&tweet.UpdatedAt,
		&tweet.LikeCount,
		&tweet.RetweetCount,
		&tweet.ReplyCount,
	); err != nil {
		return nil, err
	}

	if parentTweetID.Valid {
		tweet.ParentTweetID = &parentTweetID.Int64
	}

	return &tweet, nil
}

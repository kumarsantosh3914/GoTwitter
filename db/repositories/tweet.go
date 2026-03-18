package db

import (
	"GoTwitter/models"
	"context"
	"database/sql"
	"errors"
)

type TweetRepository interface {
	Create(ctx context.Context, tweet *models.Tweet) (*models.Tweet, error)
	GetByID(ctx context.Context, id int64) (*models.Tweet, error)
	GetAll(ctx context.Context, limit, offset int) ([]*models.Tweet, error)
	Update(ctx context.Context, tweet *models.Tweet) error
	DeleteByID(ctx context.Context, id int64) error
}

type TweetRepositoryImpl struct {
	db *sql.DB
}

func NewTweetRepository(_db *sql.DB) TweetRepository {
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
		`INSERT INTO tweets (user_id, tweet, created_at, updated_at)
		 VALUES (?, ?, NOW(), NOW())`,
		tweet.UserID,
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
	err := t.db.QueryRowContext(
		ctx,
		`SELECT id, user_id, tweet, created_at, updated_at
		 FROM tweets
		 WHERE id = ?
		 LIMIT 1`,
		id,
	).Scan(&tweet.ID, &tweet.UserID, &tweet.Tweet, &tweet.CreatedAt, &tweet.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &tweet, nil
}

func (t *TweetRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*models.Tweet, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if t.db == nil {
		return nil, errors.New("db is nil")
	}

	rows, err := t.db.QueryContext(
		ctx,
		`SELECT id, user_id, tweet, created_at, updated_at
		 FROM tweets
		 ORDER BY id DESC
		 LIMIT ? OFFSET ?`,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tweets []*models.Tweet
	for rows.Next() {
		var tweet models.Tweet
		if err := rows.Scan(&tweet.ID, &tweet.UserID, &tweet.Tweet, &tweet.CreatedAt, &tweet.UpdatedAt); err != nil {
			return nil, err
		}
		tweets = append(tweets, &tweet)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tweets, nil
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
		tweet.ID,
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

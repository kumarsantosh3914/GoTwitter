package db

import (
	"GoTwitter/models"
	"context"
	"database/sql"
	"errors"
)

type TagRepository interface {
	GetByName(ctx context.Context, name string) (*models.Tag, error)
	Create(ctx context.Context, tag *models.Tag) (*models.Tag, error)
	AssociateWithTweet(ctx context.Context, tweetID int64, tagID int64) error
	GetByTweetID(ctx context.Context, tweetID int64) ([]*models.Tag, error)
	DeleteAssociationsByTweetID(ctx context.Context, tweetID int64) error
}

type TagRepositoryImpl struct {
	db *sql.DB
}

func NewTagRepository(_db *sql.DB) TagRepository {
	return &TagRepositoryImpl{
		db: _db,
	}
}

func (t *TagRepositoryImpl) GetByName(ctx context.Context, name string) (*models.Tag, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if t.db == nil {
		return nil, errors.New("db is nil")
	}

	var tag models.Tag
	err := t.db.QueryRowContext(
		ctx,
		`SELECT id, name FROM tags WHERE name = ? LIMIT 1`,
		name,
	).Scan(&tag.Id, &tag.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

func (t *TagRepositoryImpl) Create(ctx context.Context, tag *models.Tag) (*models.Tag, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if tag == nil {
		return nil, errors.New("tag is nil")
	}
	if t.db == nil {
		return nil, errors.New("db is nil")
	}

	res, err := t.db.ExecContext(
		ctx,
		`INSERT INTO tags (name) VALUES (?)`,
		tag.Name,
	)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	tag.Id = id
	return tag, nil
}

func (t *TagRepositoryImpl) AssociateWithTweet(ctx context.Context, tweetID int64, tagID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if t.db == nil {
		return errors.New("db is nil")
	}

	_, err := t.db.ExecContext(
		ctx,
		`INSERT IGNORE INTO tweet_tags (tweet_id, tag_id) VALUES (?, ?)`,
		tweetID,
		tagID,
	)
	return err
}

func (t *TagRepositoryImpl) GetByTweetID(ctx context.Context, tweetID int64) ([]*models.Tag, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if t.db == nil {
		return nil, errors.New("db is nil")
	}

	rows, err := t.db.QueryContext(
		ctx,
		`SELECT t.id, t.name 
		 FROM tags t
		 JOIN tweet_tags tt ON t.id = tt.tag_id
		 WHERE tt.tweet_id = ?`,
		tweetID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.Id, &tag.Name); err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}
	return tags, rows.Err()
}

func (t *TagRepositoryImpl) DeleteAssociationsByTweetID(ctx context.Context, tweetID int64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if t.db == nil {
		return errors.New("db is nil")
	}

	_, err := t.db.ExecContext(ctx, `DELETE FROM tweet_tags WHERE tweet_id = ?`, tweetID)
	return err
}

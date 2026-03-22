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

	// New methods for Tag Management
	GetAll(ctx context.Context, limit, offset int) ([]*models.Tag, error)
	GetByID(ctx context.Context, id int64) (*models.Tag, error)
	GetTweetsByTagID(ctx context.Context, tagID int64, limit, offset int) ([]*models.Tweet, error)
	GetPopular(ctx context.Context, limit int) ([]map[string]interface{}, error)
	DeleteByID(ctx context.Context, id int64) error
}

type TagRepositoryImpl struct {
	db queryExecutor
}

func NewTagRepository(_db queryExecutor) TagRepository {
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

func (t *TagRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*models.Tag, error) {
	rows, err := t.db.QueryContext(ctx, `SELECT id, name FROM tags ORDER BY id DESC LIMIT ? OFFSET ?`, limit, offset)
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
	return tags, nil
}

func (t *TagRepositoryImpl) GetByID(ctx context.Context, id int64) (*models.Tag, error) {
	var tag models.Tag
	err := t.db.QueryRowContext(ctx, `SELECT id, name FROM tags WHERE id = ?`, id).Scan(&tag.Id, &tag.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

func (t *TagRepositoryImpl) GetTweetsByTagID(ctx context.Context, tagID int64, limit, offset int) ([]*models.Tweet, error) {
	rows, err := t.db.QueryContext(ctx, `
		SELECT tw.id, tw.user_id, tw.tweet, tw.created_at, tw.updated_at
		FROM tweets tw
		JOIN tweet_tags tt ON tw.id = tt.tweet_id
		WHERE tt.tag_id = ?
		ORDER BY tw.id DESC
		LIMIT ? OFFSET ?`, tagID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tweets []*models.Tweet
	for rows.Next() {
		var tw models.Tweet
		if err := rows.Scan(&tw.Id, &tw.UserId, &tw.Tweet, &tw.CreatedAt, &tw.UpdatedAt); err != nil {
			return nil, err
		}
		tweets = append(tweets, &tw)
	}
	return tweets, nil
}

func (t *TagRepositoryImpl) GetPopular(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	rows, err := t.db.QueryContext(ctx, `
		SELECT t.id, t.name, COUNT(tt.tweet_id) as usage_count
		FROM tags t
		LEFT JOIN tweet_tags tt ON t.id = tt.tag_id
		GROUP BY t.id
		ORDER BY usage_count DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id int64
		var name string
		var count int
		if err := rows.Scan(&id, &name, &count); err != nil {
			return nil, err
		}
		result = append(result, map[string]interface{}{
			"id":          id,
			"name":        name,
			"usage_count": count,
		})
	}
	return result, nil
}

func (t *TagRepositoryImpl) DeleteByID(ctx context.Context, id int64) error {
	_, err := t.db.ExecContext(ctx, `DELETE FROM tags WHERE id = ?`, id)
	return err
}

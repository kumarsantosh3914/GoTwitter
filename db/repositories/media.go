package db

import (
	"GoTwitter/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type MediaRepository interface {
	Create(ctx context.Context, attachment *models.MediaAttachment) (*models.MediaAttachment, error)
	GetOwnedByIDs(ctx context.Context, userID int64, ids []int64) ([]*models.MediaAttachment, error)
	GetByTweetID(ctx context.Context, tweetID int64) ([]*models.MediaAttachment, error)
	GetByTweetIDs(ctx context.Context, tweetIDs []int64) (map[int64][]*models.MediaAttachment, error)
	ReplaceTweetMedia(ctx context.Context, tweetID int64, userID int64, mediaIDs []int64) error
}

type MediaRepositoryImpl struct {
	db queryExecutor
}

func NewMediaRepository(db queryExecutor) MediaRepository {
	return &MediaRepositoryImpl{db: db}
}

func (m *MediaRepositoryImpl) Create(ctx context.Context, attachment *models.MediaAttachment) (*models.MediaAttachment, error) {
	res, err := m.db.ExecContext(
		ctx,
		`INSERT INTO media_attachments (user_id, tweet_id, s3_key, url, mime_type, size_bytes, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, NOW())`,
		attachment.UserId,
		attachment.TweetId,
		attachment.S3Key,
		attachment.Url,
		attachment.MimeType,
		attachment.SizeBytes,
	)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	attachment.Id = id
	return attachment, nil
}

func (m *MediaRepositoryImpl) GetOwnedByIDs(ctx context.Context, userID int64, ids []int64) ([]*models.MediaAttachment, error) {
	if len(ids) == 0 {
		return []*models.MediaAttachment{}, nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, userID)
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := m.db.QueryContext(
		ctx,
		fmt.Sprintf(`SELECT id, user_id, tweet_id, s3_key, url, mime_type, size_bytes, created_at
			FROM media_attachments
			WHERE user_id = ? AND id IN (%s)
			ORDER BY id ASC`, placeholders),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.MediaAttachment
	for rows.Next() {
		attachment, err := scanMediaAttachment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, attachment)
	}
	return items, rows.Err()
}

func (m *MediaRepositoryImpl) GetByTweetID(ctx context.Context, tweetID int64) ([]*models.MediaAttachment, error) {
	rows, err := m.db.QueryContext(
		ctx,
		`SELECT id, user_id, tweet_id, s3_key, url, mime_type, size_bytes, created_at
		 FROM media_attachments
		 WHERE tweet_id = ?
		 ORDER BY id ASC`,
		tweetID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.MediaAttachment
	for rows.Next() {
		attachment, err := scanMediaAttachment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, attachment)
	}
	return items, rows.Err()
}

func (m *MediaRepositoryImpl) GetByTweetIDs(ctx context.Context, tweetIDs []int64) (map[int64][]*models.MediaAttachment, error) {
	if len(tweetIDs) == 0 {
		return map[int64][]*models.MediaAttachment{}, nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(tweetIDs)), ",")
	args := make([]interface{}, 0, len(tweetIDs))
	for _, id := range tweetIDs {
		args = append(args, id)
	}

	rows, err := m.db.QueryContext(
		ctx,
		fmt.Sprintf(`SELECT id, user_id, tweet_id, s3_key, url, mime_type, size_bytes, created_at
			FROM media_attachments
			WHERE tweet_id IN (%s)
			ORDER BY tweet_id ASC, id ASC`, placeholders),
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make(map[int64][]*models.MediaAttachment, len(tweetIDs))
	for rows.Next() {
		attachment, err := scanMediaAttachment(rows)
		if err != nil {
			return nil, err
		}
		if attachment.TweetId == nil {
			continue
		}
		items[*attachment.TweetId] = append(items[*attachment.TweetId], attachment)
	}
	return items, rows.Err()
}

func (m *MediaRepositoryImpl) ReplaceTweetMedia(ctx context.Context, tweetID int64, userID int64, mediaIDs []int64) error {
	if _, err := m.db.ExecContext(ctx, `UPDATE media_attachments SET tweet_id = NULL WHERE tweet_id = ? AND user_id = ?`, tweetID, userID); err != nil {
		return err
	}
	if len(mediaIDs) == 0 {
		return nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(mediaIDs)), ",")
	args := make([]interface{}, 0, len(mediaIDs)+2)
	args = append(args, tweetID, userID)
	for _, id := range mediaIDs {
		args = append(args, id)
	}

	_, err := m.db.ExecContext(
		ctx,
		fmt.Sprintf(`UPDATE media_attachments
			SET tweet_id = ?
			WHERE user_id = ? AND id IN (%s)`, placeholders),
		args...,
	)
	return err
}

func scanMediaAttachment(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.MediaAttachment, error) {
	var attachment models.MediaAttachment
	var tweetID sql.NullInt64

	if err := scanner.Scan(
		&attachment.Id,
		&attachment.UserId,
		&tweetID,
		&attachment.S3Key,
		&attachment.Url,
		&attachment.MimeType,
		&attachment.SizeBytes,
		&attachment.CreatedAt,
	); err != nil {
		return nil, err
	}

	if tweetID.Valid {
		attachment.TweetId = &tweetID.Int64
	}

	return &attachment, nil
}

var _ = errors.New

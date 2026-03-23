package services

import (
	db "GoTwitter/db/repositories"
	apperrors "GoTwitter/errors"
	"GoTwitter/models"
	"context"
	"net/http"
)

type TagService interface {
	ListTags(ctx context.Context, page, pageSize int) ([]*models.Tag, error)
	GetTagWithTweets(ctx context.Context, id int64, page, pageSize int) (*models.Tag, []*models.Tweet, error)
	GetPopularTags(ctx context.Context, limit int) ([]map[string]interface{}, error)
	DeleteTag(ctx context.Context, id int64) error
}

type TagServiceImpl struct {
	tagRepository db.TagRepository
}

func NewTagService(_tagRepository db.TagRepository) TagService {
	return &TagServiceImpl{
		tagRepository: _tagRepository,
	}
}

func (s *TagServiceImpl) ListTags(ctx context.Context, page, pageSize int) ([]*models.Tag, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	tags, err := s.tagRepository.GetAll(ctx, pageSize, offset)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch tags", http.StatusInternalServerError, err)
	}
	return tags, nil
}

func (s *TagServiceImpl) GetTagWithTweets(ctx context.Context, id int64, page, pageSize int) (*models.Tag, []*models.Tweet, error) {
	tag, err := s.tagRepository.GetByID(ctx, id)
	if err != nil {
		return nil, nil, apperrors.NewAppError("failed to fetch tag", http.StatusInternalServerError, err)
	}
	if tag == nil {
		return nil, nil, apperrors.NewAppError("tag not found", http.StatusNotFound, nil)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	tweets, err := s.tagRepository.GetTweetsByTagID(ctx, id, pageSize, offset)
	if err != nil {
		return nil, nil, apperrors.NewAppError("failed to fetch tweets for tag", http.StatusInternalServerError, err)
	}

	return tag, tweets, nil
}

func (s *TagServiceImpl) GetPopularTags(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	if limit < 1 {
		limit = 10
	}
	tags, err := s.tagRepository.GetPopular(ctx, limit)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch popular tags", http.StatusInternalServerError, err)
	}
	return tags, nil
}

func (s *TagServiceImpl) DeleteTag(ctx context.Context, id int64) error {
	tag, err := s.tagRepository.GetByID(ctx, id)
	if err != nil {
		return apperrors.NewAppError("failed to fetch tag", http.StatusInternalServerError, err)
	}
	if tag == nil {
		return apperrors.NewAppError("tag not found", http.StatusNotFound, nil)
	}

	if err := s.tagRepository.DeleteByID(ctx, id); err != nil {
		return apperrors.NewAppError("failed to delete tag", http.StatusInternalServerError, err)
	}
	return nil
}

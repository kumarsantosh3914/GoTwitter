package services

import (
	db "GoTwitter/db/repositories"
	apperrors "GoTwitter/errors"
	"GoTwitter/models"
	"context"
	"net/http"
)

type TweetService interface {
	CreateTweet(ctx context.Context, tweet *models.Tweet) (*models.Tweet, error)
	ListTweets(ctx context.Context, page, pageSize int) ([]*models.Tweet, error)
	GetTweetByID(ctx context.Context, id int64) (*models.Tweet, error)
	UpdateTweet(ctx context.Context, tweet *models.Tweet) error
	DeleteTweet(ctx context.Context, id int64, userId int64) error
}

type TweetServiceImpl struct {
	tweetRepository db.TweetRepository
}

func NewTweetService(_tweetRepository db.TweetRepository) TweetService {
	return &TweetServiceImpl{
		tweetRepository: _tweetRepository,
	}
}

func (t *TweetServiceImpl) CreateTweet(ctx context.Context, tweet *models.Tweet) (*models.Tweet, error) {
	if tweet == nil {
		return nil, apperrors.NewAppError("tweet is nil", http.StatusBadRequest, nil)
	}
	if tweet.Tweet == "" {
		return nil, apperrors.NewAppError("tweet content is required", http.StatusBadRequest, nil)
	}

	createdTweet, err := t.tweetRepository.Create(ctx, tweet)
	if err != nil {
		return nil, apperrors.NewAppError("failed to create tweet", http.StatusInternalServerError, err)
	}

	return createdTweet, nil
}

func (t *TweetServiceImpl) ListTweets(ctx context.Context, page, pageSize int) ([]*models.Tweet, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	tweets, err := t.tweetRepository.GetAll(ctx, pageSize, offset)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch tweets", http.StatusInternalServerError, err)
	}
	return tweets, nil
}

func (t *TweetServiceImpl) GetTweetByID(ctx context.Context, id int64) (*models.Tweet, error) {
	tweet, err := t.tweetRepository.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch tweet", http.StatusInternalServerError, err)
	}
	if tweet == nil {
		return nil, apperrors.NewAppError("tweet not found", http.StatusNotFound, nil)
	}
	return tweet, nil
}

func (t *TweetServiceImpl) UpdateTweet(ctx context.Context, tweet *models.Tweet) error {
	if tweet == nil {
		return apperrors.NewAppError("tweet is nil", http.StatusBadRequest, nil)
	}

	// Check if tweet exists
	existing, err := t.tweetRepository.GetByID(ctx, int64(tweet.ID))
	if err != nil {
		return apperrors.NewAppError("failed to fetch tweet", http.StatusInternalServerError, err)
	}
	if existing == nil {
		return apperrors.NewAppError("tweet not found", http.StatusNotFound, nil)
	}

	// Only author can update
	if existing.UserID != tweet.UserID {
		return apperrors.NewAppError("unauthorized: only the author can update the tweet", http.StatusForbidden, nil)
	}

	if err := t.tweetRepository.Update(ctx, tweet); err != nil {
		return apperrors.NewAppError("failed to update tweet", http.StatusInternalServerError, err)
	}
	return nil
}

func (t *TweetServiceImpl) DeleteTweet(ctx context.Context, id int64, userId int64) error {
	// Check if tweet exists
	existing, err := t.tweetRepository.GetByID(ctx, id)
	if err != nil {
		return apperrors.NewAppError("failed to fetch tweet", http.StatusInternalServerError, err)
	}
	if existing == nil {
		return apperrors.NewAppError("tweet not found", http.StatusNotFound, nil)
	}

	// Only author can delete
	if existing.UserID != uint64(userId) {
		return apperrors.NewAppError("unauthorized: only the author can delete the tweet", http.StatusForbidden, nil)
	}

	if err := t.tweetRepository.DeleteByID(ctx, id); err != nil {
		return apperrors.NewAppError("failed to delete tweet", http.StatusInternalServerError, err)
	}
	return nil
}

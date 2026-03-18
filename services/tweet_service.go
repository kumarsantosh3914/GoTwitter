package services

import (
	apperrors "GoTwitter/errors"
	db "GoTwitter/db/repositories"
	"GoTwitter/models"
	"context"
	"net/http"
	"regexp"
	"strings"
)

type TweetService interface {
	CreateTweet(ctx context.Context, tweet *models.Tweet) (*models.Tweet, error)
	ListTweets(ctx context.Context, page, pageSize int, userId int64, tag string, search string) ([]*models.Tweet, error)
	GetTweetByID(ctx context.Context, id int64) (*models.Tweet, error)
	UpdateTweet(ctx context.Context, tweet *models.Tweet) error
	DeleteTweet(ctx context.Context, id int64, userId int64) error
}

type TweetServiceImpl struct {
	tweetRepository db.TweetRepository
	tagRepository   db.TagRepository
}

func NewTweetService(_tweetRepository db.TweetRepository, _tagRepository db.TagRepository) TweetService {
	return &TweetServiceImpl{
		tweetRepository: _tweetRepository,
		tagRepository:   _tagRepository,
	}
}

func (t *TweetServiceImpl) CreateTweet(ctx context.Context, tweet *models.Tweet) (*models.Tweet, error) {
	if tweet == nil {
		return nil, apperrors.NewAppError("tweet is nil", http.StatusBadRequest, nil)
	}
	if tweet.Tweet == "" {
		return nil, apperrors.NewAppError("tweet content is required", http.StatusBadRequest, nil)
	}
	if len(tweet.Tweet) > 280 {
		return nil, apperrors.NewAppError("tweet content exceeds 280 characters", http.StatusBadRequest, nil)
	}

	createdTweet, err := t.tweetRepository.Create(ctx, tweet)
	if err != nil {
		return nil, apperrors.NewAppError("failed to create tweet", http.StatusInternalServerError, err)
	}

	// Extract and associate hashtags
	hashtags := extractHashtags(tweet.Tweet)
	for _, h := range hashtags {
		tag, err := t.tagRepository.GetByName(ctx, h)
		if err != nil {
			continue
		}
		if tag == nil {
			tag, err = t.tagRepository.Create(ctx, &models.Tag{Name: h})
			if err != nil {
				continue
			}
		}
		_ = t.tagRepository.AssociateWithTweet(ctx, createdTweet.Id, tag.Id)
	}

	// Fetch tags to include in response
	tags, _ := t.tagRepository.GetByTweetID(ctx, createdTweet.Id)
	createdTweet.Tags = tags

	return createdTweet, nil
}

func (t *TweetServiceImpl) ListTweets(ctx context.Context, page, pageSize int, userId int64, tag string, search string) ([]*models.Tweet, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	tweets, err := t.tweetRepository.GetAll(ctx, pageSize, offset, userId, tag, search)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch tweets", http.StatusInternalServerError, err)
	}

	for _, tweet := range tweets {
		tags, _ := t.tagRepository.GetByTweetID(ctx, tweet.Id)
		tweet.Tags = tags
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

	tags, _ := t.tagRepository.GetByTweetID(ctx, tweet.Id)
	tweet.Tags = tags

	return tweet, nil
}

func (t *TweetServiceImpl) UpdateTweet(ctx context.Context, tweet *models.Tweet) error {
	if tweet == nil {
		return apperrors.NewAppError("tweet is nil", http.StatusBadRequest, nil)
	}
	if len(tweet.Tweet) > 280 {
		return apperrors.NewAppError("tweet content exceeds 280 characters", http.StatusBadRequest, nil)
	}

	// Check if tweet exists
	existing, err := t.tweetRepository.GetByID(ctx, tweet.Id)
	if err != nil {
		return apperrors.NewAppError("failed to fetch tweet", http.StatusInternalServerError, err)
	}
	if existing == nil {
		return apperrors.NewAppError("tweet not found", http.StatusNotFound, nil)
	}

	// Only author can update
	if existing.UserId != tweet.UserId {
		return apperrors.NewAppError("unauthorized: only the author can update the tweet", http.StatusForbidden, nil)
	}

	if err := t.tweetRepository.Update(ctx, tweet); err != nil {
		return apperrors.NewAppError("failed to update tweet", http.StatusInternalServerError, err)
	}

	// Update hashtag associations: clear old ones and add new ones
	_ = t.tagRepository.DeleteAssociationsByTweetID(ctx, tweet.Id)
	hashtags := extractHashtags(tweet.Tweet)
	for _, h := range hashtags {
		tag, err := t.tagRepository.GetByName(ctx, h)
		if err != nil {
			continue
		}
		if tag == nil {
			tag, err = t.tagRepository.Create(ctx, &models.Tag{Name: h})
			if err != nil {
				continue
			}
		}
		_ = t.tagRepository.AssociateWithTweet(ctx, tweet.Id, tag.Id)
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
	if existing.UserId != userId {
		return apperrors.NewAppError("unauthorized: only the author can delete the tweet", http.StatusForbidden, nil)
	}

	if err := t.tweetRepository.DeleteByID(ctx, id); err != nil {
		return apperrors.NewAppError("failed to delete tweet", http.StatusInternalServerError, err)
	}
	return nil
}

func extractHashtags(text string) []string {
	re := regexp.MustCompile(`#[a-zA-Z0-9_]+`)
	matches := re.FindAllString(text, -1)
	
	uniqueTags := make(map[string]bool)
	var tags []string
	
	for _, match := range matches {
		tag := strings.ToLower(strings.TrimPrefix(match, "#"))
		if !uniqueTags[tag] {
			uniqueTags[tag] = true
			tags = append(tags, tag)
		}
	}
	
	return tags
}

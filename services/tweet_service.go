package services

import (
	db "GoTwitter/db/repositories"
	apperrors "GoTwitter/errors"
	"GoTwitter/models"
	"context"
	"database/sql"
	"net/http"
	"regexp"
	"strings"
)

type TweetService interface {
	CreateTweet(ctx context.Context, tweet *models.Tweet, mediaIDs []int64) (*models.Tweet, error)
	ListTweets(ctx context.Context, page int, pageSize int, userId int64, tag string, search string, viewerID *int64) ([]*models.Tweet, error)
	GetTweetByID(ctx context.Context, id int64, viewerID *int64) (*models.Tweet, error)
	GetThread(ctx context.Context, id int64, viewerID *int64) (*models.Tweet, error)
	UpdateTweet(ctx context.Context, tweet *models.Tweet, mediaIDs []int64) error
	DeleteTweet(ctx context.Context, id int64, userId int64) error
	LikeTweet(ctx context.Context, tweetID int64, userID int64) (*models.Tweet, error)
	UnlikeTweet(ctx context.Context, tweetID int64, userID int64) (*models.Tweet, error)
	RetweetTweet(ctx context.Context, tweetID int64, userID int64) (*models.Tweet, error)
	UnretweetTweet(ctx context.Context, tweetID int64, userID int64) (*models.Tweet, error)
}

type TweetServiceImpl struct {
	db               *sql.DB
	tweetRepository  db.TweetRepository
	tagRepository    db.TagRepository
	socialRepository db.SocialRepository
	mediaRepository  db.MediaRepository
}

func NewTweetService(dbConn *sql.DB, tweetRepository db.TweetRepository, tagRepository db.TagRepository, socialRepository db.SocialRepository, mediaRepository db.MediaRepository) TweetService {
	return &TweetServiceImpl{
		db:               dbConn,
		tweetRepository:  tweetRepository,
		tagRepository:    tagRepository,
		socialRepository: socialRepository,
		mediaRepository:  mediaRepository,
	}
}

func (t *TweetServiceImpl) CreateTweet(ctx context.Context, tweet *models.Tweet, mediaIDs []int64) (*models.Tweet, error) {
	if tweet == nil {
		return nil, apperrors.NewAppError("tweet is nil", http.StatusBadRequest, nil)
	}
	if tweet.Tweet == "" {
		return nil, apperrors.NewAppError("tweet content is required", http.StatusBadRequest, nil)
	}
	if len(tweet.Tweet) > 280 {
		return nil, apperrors.NewAppError("tweet content exceeds 280 characters", http.StatusBadRequest, nil)
	}
	if len(mediaIDs) > 4 {
		return nil, apperrors.NewAppError("a tweet can include at most 4 media attachments", http.StatusBadRequest, nil)
	}

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, apperrors.NewAppError("failed to start tweet transaction", http.StatusInternalServerError, err)
	}
	defer tx.Rollback()

	txTweetRepository := db.NewTweetRepository(tx)
	txTagRepository := db.NewTagRepository(tx)
	txMediaRepository := db.NewMediaRepository(tx)

	if tweet.ParentTweetID != nil {
		parent, err := txTweetRepository.GetByID(ctx, *tweet.ParentTweetID)
		if err != nil {
			return nil, apperrors.NewAppError("failed to fetch parent tweet", http.StatusInternalServerError, err)
		}
		if parent == nil {
			return nil, apperrors.NewAppError("parent tweet not found", http.StatusNotFound, nil)
		}
	}

	if err := t.validateOwnedMedia(ctx, txMediaRepository, tweet.UserId, mediaIDs); err != nil {
		return nil, err
	}

	createdTweet, err := txTweetRepository.Create(ctx, tweet)
	if err != nil {
		return nil, apperrors.NewAppError("failed to create tweet", http.StatusInternalServerError, err)
	}

	if err := syncTweetTags(ctx, txTagRepository, createdTweet.Id, tweet.Tweet); err != nil {
		return nil, err
	}
	if err := txMediaRepository.ReplaceTweetMedia(ctx, createdTweet.Id, tweet.UserId, mediaIDs); err != nil {
		return nil, apperrors.NewAppError("failed to attach media to tweet", http.StatusInternalServerError, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, apperrors.NewAppError("failed to commit tweet transaction", http.StatusInternalServerError, err)
	}

	return t.GetTweetByID(ctx, createdTweet.Id, &tweet.UserId)
}

func (t *TweetServiceImpl) ListTweets(ctx context.Context, page int, pageSize int, userId int64, tag string, search string, viewerID *int64) ([]*models.Tweet, error) {
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

	if err := t.enrichTweets(ctx, tweets, viewerID); err != nil {
		return nil, err
	}

	return tweets, nil
}

func (t *TweetServiceImpl) GetTweetByID(ctx context.Context, id int64, viewerID *int64) (*models.Tweet, error) {
	tweet, err := t.tweetRepository.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch tweet", http.StatusInternalServerError, err)
	}
	if tweet == nil {
		return nil, apperrors.NewAppError("tweet not found", http.StatusNotFound, nil)
	}

	if err := t.enrichTweets(ctx, []*models.Tweet{tweet}, viewerID); err != nil {
		return nil, err
	}
	if err := t.attachReplyTree(ctx, tweet, viewerID); err != nil {
		return nil, err
	}

	return tweet, nil
}

func (t *TweetServiceImpl) GetThread(ctx context.Context, id int64, viewerID *int64) (*models.Tweet, error) {
	tweet, err := t.GetTweetByID(ctx, id, viewerID)
	if err != nil {
		return nil, err
	}

	var chain []*models.Tweet
	current := tweet
	for {
		chain = append([]*models.Tweet{current}, chain...)
		if current.ParentTweetID == nil {
			break
		}
		parent, err := t.tweetRepository.GetByID(ctx, *current.ParentTweetID)
		if err != nil {
			return nil, apperrors.NewAppError("failed to fetch thread", http.StatusInternalServerError, err)
		}
		if parent == nil {
			break
		}
		current = parent
	}

	if err := t.enrichTweets(ctx, chain, viewerID); err != nil {
		return nil, err
	}
	tweet.Thread = chain
	return tweet, nil
}

func (t *TweetServiceImpl) UpdateTweet(ctx context.Context, tweet *models.Tweet, mediaIDs []int64) error {
	if tweet == nil {
		return apperrors.NewAppError("tweet is nil", http.StatusBadRequest, nil)
	}
	if len(tweet.Tweet) > 280 {
		return apperrors.NewAppError("tweet content exceeds 280 characters", http.StatusBadRequest, nil)
	}
	if len(mediaIDs) > 4 {
		return apperrors.NewAppError("a tweet can include at most 4 media attachments", http.StatusBadRequest, nil)
	}

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return apperrors.NewAppError("failed to start tweet transaction", http.StatusInternalServerError, err)
	}
	defer tx.Rollback()

	txTweetRepository := db.NewTweetRepository(tx)
	txTagRepository := db.NewTagRepository(tx)
	txMediaRepository := db.NewMediaRepository(tx)

	existing, err := txTweetRepository.GetByID(ctx, tweet.Id)
	if err != nil {
		return apperrors.NewAppError("failed to fetch tweet", http.StatusInternalServerError, err)
	}
	if existing == nil {
		return apperrors.NewAppError("tweet not found", http.StatusNotFound, nil)
	}
	if existing.UserId != tweet.UserId {
		return apperrors.NewAppError("unauthorized: only the author can update the tweet", http.StatusForbidden, nil)
	}
	tweet.ParentTweetID = existing.ParentTweetID

	if err := t.validateOwnedMedia(ctx, txMediaRepository, tweet.UserId, mediaIDs); err != nil {
		return err
	}
	if err := txTweetRepository.Update(ctx, tweet); err != nil {
		return apperrors.NewAppError("failed to update tweet", http.StatusInternalServerError, err)
	}
	if err := txTagRepository.DeleteAssociationsByTweetID(ctx, tweet.Id); err != nil {
		return apperrors.NewAppError("failed to clear tweet tag associations", http.StatusInternalServerError, err)
	}
	if err := syncTweetTags(ctx, txTagRepository, tweet.Id, tweet.Tweet); err != nil {
		return err
	}
	if err := txMediaRepository.ReplaceTweetMedia(ctx, tweet.Id, tweet.UserId, mediaIDs); err != nil {
		return apperrors.NewAppError("failed to update media attachments", http.StatusInternalServerError, err)
	}

	if err := tx.Commit(); err != nil {
		return apperrors.NewAppError("failed to commit tweet transaction", http.StatusInternalServerError, err)
	}

	return nil
}

func (t *TweetServiceImpl) DeleteTweet(ctx context.Context, id int64, userId int64) error {
	existing, err := t.tweetRepository.GetByID(ctx, id)
	if err != nil {
		return apperrors.NewAppError("failed to fetch tweet", http.StatusInternalServerError, err)
	}
	if existing == nil {
		return apperrors.NewAppError("tweet not found", http.StatusNotFound, nil)
	}
	if existing.UserId != userId {
		return apperrors.NewAppError("unauthorized: only the author can delete the tweet", http.StatusForbidden, nil)
	}
	if err := t.tweetRepository.DeleteByID(ctx, id); err != nil {
		return apperrors.NewAppError("failed to delete tweet", http.StatusInternalServerError, err)
	}
	return nil
}

func (t *TweetServiceImpl) LikeTweet(ctx context.Context, tweetID int64, userID int64) (*models.Tweet, error) {
	tweet, err := t.tweetRepository.GetByID(ctx, tweetID)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch tweet", http.StatusInternalServerError, err)
	}
	if tweet == nil {
		return nil, apperrors.NewAppError("tweet not found", http.StatusNotFound, nil)
	}
	if err := t.socialRepository.LikeTweet(ctx, tweetID, userID); err != nil {
		return nil, apperrors.NewAppError("failed to like tweet", http.StatusInternalServerError, err)
	}
	return t.GetTweetByID(ctx, tweetID, &userID)
}

func (t *TweetServiceImpl) UnlikeTweet(ctx context.Context, tweetID int64, userID int64) (*models.Tweet, error) {
	tweet, err := t.tweetRepository.GetByID(ctx, tweetID)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch tweet", http.StatusInternalServerError, err)
	}
	if tweet == nil {
		return nil, apperrors.NewAppError("tweet not found", http.StatusNotFound, nil)
	}
	if err := t.socialRepository.UnlikeTweet(ctx, tweetID, userID); err != nil {
		return nil, apperrors.NewAppError("failed to unlike tweet", http.StatusInternalServerError, err)
	}
	return t.GetTweetByID(ctx, tweetID, &userID)
}

func (t *TweetServiceImpl) RetweetTweet(ctx context.Context, tweetID int64, userID int64) (*models.Tweet, error) {
	tweet, err := t.tweetRepository.GetByID(ctx, tweetID)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch tweet", http.StatusInternalServerError, err)
	}
	if tweet == nil {
		return nil, apperrors.NewAppError("tweet not found", http.StatusNotFound, nil)
	}
	if err := t.socialRepository.RetweetTweet(ctx, tweetID, userID); err != nil {
		return nil, apperrors.NewAppError("failed to retweet tweet", http.StatusInternalServerError, err)
	}
	return t.GetTweetByID(ctx, tweetID, &userID)
}

func (t *TweetServiceImpl) UnretweetTweet(ctx context.Context, tweetID int64, userID int64) (*models.Tweet, error) {
	tweet, err := t.tweetRepository.GetByID(ctx, tweetID)
	if err != nil {
		return nil, apperrors.NewAppError("failed to fetch tweet", http.StatusInternalServerError, err)
	}
	if tweet == nil {
		return nil, apperrors.NewAppError("tweet not found", http.StatusNotFound, nil)
	}
	if err := t.socialRepository.UnretweetTweet(ctx, tweetID, userID); err != nil {
		return nil, apperrors.NewAppError("failed to remove retweet", http.StatusInternalServerError, err)
	}
	return t.GetTweetByID(ctx, tweetID, &userID)
}

func (t *TweetServiceImpl) validateOwnedMedia(ctx context.Context, mediaRepository db.MediaRepository, userID int64, mediaIDs []int64) error {
	if len(mediaIDs) == 0 {
		return nil
	}
	items, err := mediaRepository.GetOwnedByIDs(ctx, userID, mediaIDs)
	if err != nil {
		return apperrors.NewAppError("failed to fetch media attachments", http.StatusInternalServerError, err)
	}
	if len(items) != len(mediaIDs) {
		return apperrors.NewAppError("one or more media attachments do not belong to the current user", http.StatusBadRequest, nil)
	}
	return nil
}

func (t *TweetServiceImpl) enrichTweets(ctx context.Context, tweets []*models.Tweet, viewerID *int64) error {
	if len(tweets) == 0 {
		return nil
	}

	tweetIDs := make([]int64, 0, len(tweets))
	for _, tweet := range tweets {
		tweetIDs = append(tweetIDs, tweet.Id)
	}

	tagsByTweetID, err := t.tagRepository.GetByTweetIDs(ctx, tweetIDs)
	if err != nil {
		return apperrors.NewAppError("failed to fetch tweet tags", http.StatusInternalServerError, err)
	}
	mediaByTweetID, err := t.mediaRepository.GetByTweetIDs(ctx, tweetIDs)
	if err != nil {
		return apperrors.NewAppError("failed to fetch tweet media", http.StatusInternalServerError, err)
	}

	var interactionStates map[int64]db.TweetInteractionState
	if viewerID != nil {
		interactionStates, err = t.socialRepository.GetTweetInteractionStates(ctx, *viewerID, tweetIDs)
		if err != nil {
			return apperrors.NewAppError("failed to fetch tweet interactions", http.StatusInternalServerError, err)
		}
	}

	for _, tweet := range tweets {
		tweet.Tags = tagsByTweetID[tweet.Id]
		tweet.Media = mediaByTweetID[tweet.Id]
		if interactionStates != nil {
			state := interactionStates[tweet.Id]
			tweet.IsLiked = state.IsLiked
			tweet.IsRetweeted = state.IsRetweeted
		}
	}

	return nil
}

func (t *TweetServiceImpl) attachReplyTree(ctx context.Context, tweet *models.Tweet, viewerID *int64) error {
	parentIDs := []int64{tweet.Id}
	replyMaps := make(map[int64][]*models.Tweet)

	for len(parentIDs) > 0 {
		currentReplies, err := t.tweetRepository.GetRepliesByParentIDs(ctx, parentIDs)
		if err != nil {
			return apperrors.NewAppError("failed to fetch tweet replies", http.StatusInternalServerError, err)
		}

		var batch []*models.Tweet
		var nextParentIDs []int64
		for _, parentID := range parentIDs {
			replies := currentReplies[parentID]
			if len(replies) == 0 {
				continue
			}
			replyMaps[parentID] = replies
			for _, reply := range replies {
				batch = append(batch, reply)
				nextParentIDs = append(nextParentIDs, reply.Id)
			}
		}

		if err := t.enrichTweets(ctx, batch, viewerID); err != nil {
			return err
		}
		parentIDs = nextParentIDs
	}

	var assignReplies func(node *models.Tweet)
	assignReplies = func(node *models.Tweet) {
		node.Replies = replyMaps[node.Id]
		for _, reply := range node.Replies {
			assignReplies(reply)
		}
	}

	assignReplies(tweet)
	return nil
}

func syncTweetTags(ctx context.Context, tagRepository db.TagRepository, tweetID int64, content string) error {
	hashtags := extractHashtags(content)
	for _, h := range hashtags {
		tag, err := tagRepository.GetByName(ctx, h)
		if err != nil {
			return apperrors.NewAppError("failed to fetch tag", http.StatusInternalServerError, err)
		}
		if tag == nil {
			tag, err = tagRepository.Create(ctx, &models.Tag{Name: h})
			if err != nil {
				return apperrors.NewAppError("failed to create tag", http.StatusInternalServerError, err)
			}
		}
		if err := tagRepository.AssociateWithTweet(ctx, tweetID, tag.Id); err != nil {
			return apperrors.NewAppError("failed to associate tag with tweet", http.StatusInternalServerError, err)
		}
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

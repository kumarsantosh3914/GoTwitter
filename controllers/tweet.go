package controllers

import (
	apperrors "GoTwitter/errors"
	"GoTwitter/models"
	"GoTwitter/services"
	"GoTwitter/utils"
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type TweetController struct {
	TweetService services.TweetService
}

func NewTweetController(_tweetService services.TweetService) *TweetController {
	return &TweetController{
		TweetService: _tweetService,
	}
}

func (tc *TweetController) CreateTweet(w http.ResponseWriter, r *http.Request) {
	claims, ok := utils.GetUserFromContext(r.Context())
	if !ok {
		handleError(w, apperrors.NewAppError("unauthorized", http.StatusUnauthorized, nil))
		return
	}

	var payload struct {
		Tweet         string  `json:"tweet" validate:"required,max=280"`
		ParentTweetID *int64  `json:"parent_tweet_id"`
		MediaIDs      []int64 `json:"media_ids"`
	}

	if err := utils.ReadJsonBody(r, &payload); err != nil {
		handleError(w, apperrors.NewAppError("invalid json body", http.StatusBadRequest, err))
		return
	}

	if err := utils.Validator.Struct(payload); err != nil {
		handleError(w, apperrors.NewAppError("validation failed", http.StatusBadRequest, err))
		return
	}

	created, err := tc.TweetService.CreateTweet(r.Context(), &models.Tweet{
		UserId:        claims.UserID,
		ParentTweetID: payload.ParentTweetID,
		Tweet:         payload.Tweet,
	}, payload.MediaIDs)
	if err != nil {
		log.Printf("[ERROR] create tweet failed: %v", err)
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusCreated, "Tweet created successfully", created)
}

func (tc *TweetController) ListTweets(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, err := parsePositiveIntQuery(query.Get("page"), "page", 1)
	if err != nil {
		handleError(w, err)
		return
	}

	pageSize, err := parsePositiveIntQuery(query.Get("page_size"), "page_size", 10)
	if err != nil {
		handleError(w, err)
		return
	}

	userId, err := parsePositiveInt64Query(query.Get("user_id"), "user_id")
	if err != nil {
		handleError(w, err)
		return
	}
	tag := query.Get("tag")
	search := query.Get("q")
	var viewerID *int64
	if claims, ok := utils.GetUserFromRequest(r); ok {
		viewerID = &claims.UserID
	}

	tweets, err := tc.TweetService.ListTweets(r.Context(), page, pageSize, userId, tag, search, viewerID)
	if err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Tweets fetched successfully", map[string]any{
		"items": tweets,
		"meta": paginationMeta{
			Page:     page,
			PageSize: pageSize,
			Count:    len(tweets),
		},
	})
}

func (tc *TweetController) GetTweet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handleError(w, apperrors.NewAppError("invalid tweet id", http.StatusBadRequest, err))
		return
	}

	var viewerID *int64
	if claims, ok := utils.GetUserFromRequest(r); ok {
		viewerID = &claims.UserID
	}

	tweet, err := tc.TweetService.GetTweetByID(r.Context(), id, viewerID)
	if err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Tweet fetched successfully", tweet)
}

func (tc *TweetController) GetThread(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handleError(w, apperrors.NewAppError("invalid tweet id", http.StatusBadRequest, err))
		return
	}

	var viewerID *int64
	if claims, ok := utils.GetUserFromRequest(r); ok {
		viewerID = &claims.UserID
	}

	tweet, err := tc.TweetService.GetThread(r.Context(), id, viewerID)
	if err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Thread fetched successfully", tweet)
}

func (tc *TweetController) UpdateTweet(w http.ResponseWriter, r *http.Request) {
	claims, ok := utils.GetUserFromContext(r.Context())
	if !ok {
		handleError(w, apperrors.NewAppError("unauthorized", http.StatusUnauthorized, nil))
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handleError(w, apperrors.NewAppError("invalid tweet id", http.StatusBadRequest, err))
		return
	}

	var payload struct {
		Tweet    string  `json:"tweet" validate:"required,max=280"`
		MediaIDs []int64 `json:"media_ids"`
	}

	if err := utils.ReadJsonBody(r, &payload); err != nil {
		handleError(w, apperrors.NewAppError("invalid json body", http.StatusBadRequest, err))
		return
	}

	if err := utils.Validator.Struct(payload); err != nil {
		handleError(w, apperrors.NewAppError("validation failed", http.StatusBadRequest, err))
		return
	}

	tweet := &models.Tweet{
		Id:     id,
		UserId: claims.UserID,
		Tweet:  payload.Tweet,
	}

	if err := tc.TweetService.UpdateTweet(r.Context(), tweet, payload.MediaIDs); err != nil {
		handleError(w, err)
		return
	}

	// Fetch updated tweet to return
	updated, err := tc.TweetService.GetTweetByID(r.Context(), id, &claims.UserID)
	if err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Tweet updated successfully", updated)
}

func (tc *TweetController) DeleteTweet(w http.ResponseWriter, r *http.Request) {
	claims, ok := utils.GetUserFromContext(r.Context())
	if !ok {
		handleError(w, apperrors.NewAppError("unauthorized", http.StatusUnauthorized, nil))
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handleError(w, apperrors.NewAppError("invalid tweet id", http.StatusBadRequest, err))
		return
	}

	if err := tc.TweetService.DeleteTweet(r.Context(), id, claims.UserID); err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Tweet deleted successfully", nil)
}

func (tc *TweetController) LikeTweet(w http.ResponseWriter, r *http.Request) {
	tc.toggleTweetInteraction(w, r, tc.TweetService.LikeTweet, "Tweet liked successfully")
}

func (tc *TweetController) UnlikeTweet(w http.ResponseWriter, r *http.Request) {
	tc.toggleTweetInteraction(w, r, tc.TweetService.UnlikeTweet, "Tweet unliked successfully")
}

func (tc *TweetController) RetweetTweet(w http.ResponseWriter, r *http.Request) {
	tc.toggleTweetInteraction(w, r, tc.TweetService.RetweetTweet, "Tweet retweeted successfully")
}

func (tc *TweetController) UnretweetTweet(w http.ResponseWriter, r *http.Request) {
	tc.toggleTweetInteraction(w, r, tc.TweetService.UnretweetTweet, "Tweet unretweeted successfully")
}

func (tc *TweetController) toggleTweetInteraction(
	w http.ResponseWriter,
	r *http.Request,
	action func(context.Context, int64, int64) (*models.Tweet, error),
	message string,
) {
	claims, ok := utils.GetUserFromContext(r.Context())
	if !ok {
		handleError(w, apperrors.NewAppError("unauthorized", http.StatusUnauthorized, nil))
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handleError(w, apperrors.NewAppError("invalid tweet id", http.StatusBadRequest, err))
		return
	}

	tweet, err := action(r.Context(), id, claims.UserID)
	if err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, message, tweet)
}

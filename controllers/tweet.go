package controllers

import (
	apperrors "GoTwitter/errors"
	"GoTwitter/models"
	"GoTwitter/services"
	"GoTwitter/utils"
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
		Tweet string `json:"tweet" validate:"required,max=280"`
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
		UserId: claims.UserID,
		Tweet:  payload.Tweet,
	})
	if err != nil {
		log.Printf("[ERROR] create tweet failed: %v", err)
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusCreated, "Tweet created successfully", created)
}

func (tc *TweetController) ListTweets(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	
	userId, _ := strconv.ParseInt(query.Get("user_id"), 10, 64)
	tag := query.Get("tag")
	search := query.Get("q")

	tweets, err := tc.TweetService.ListTweets(r.Context(), page, pageSize, userId, tag, search)
	if err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Tweets fetched successfully", tweets)
}

func (tc *TweetController) GetTweet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handleError(w, apperrors.NewAppError("invalid tweet id", http.StatusBadRequest, err))
		return
	}

	tweet, err := tc.TweetService.GetTweetByID(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Tweet fetched successfully", tweet)
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
		Tweet string `json:"tweet" validate:"required,max=280"`
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

	if err := tc.TweetService.UpdateTweet(r.Context(), tweet); err != nil {
		handleError(w, err)
		return
	}

	// Fetch updated tweet to return
	updated, _ := tc.TweetService.GetTweetByID(r.Context(), id)

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

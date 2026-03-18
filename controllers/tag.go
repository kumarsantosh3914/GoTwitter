package controllers

import (
	apperrors "GoTwitter/errors"
	"GoTwitter/services"
	"GoTwitter/utils"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type TagController struct {
	TagService services.TagService
}

func NewTagController(_tagService services.TagService) *TagController {
	return &TagController{
		TagService: _tagService,
	}
}

func (tc *TagController) ListTags(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	tags, err := tc.TagService.ListTags(r.Context(), page, pageSize)
	if err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Tags fetched successfully", tags)
}

func (tc *TagController) GetTagDetails(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handleError(w, apperrors.NewAppError("invalid tag id", http.StatusBadRequest, err))
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	tag, tweets, err := tc.TagService.GetTagWithTweets(r.Context(), id, page, pageSize)
	if err != nil {
		handleError(w, err)
		return
	}

	response := map[string]interface{}{
		"tag":    tag,
		"tweets": tweets,
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Tag details fetched successfully", response)
}

func (tc *TagController) GetPopularTags(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	tags, err := tc.TagService.GetPopularTags(r.Context(), limit)
	if err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Popular tags fetched successfully", tags)
}

func (tc *TagController) DeleteTag(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handleError(w, apperrors.NewAppError("invalid tag id", http.StatusBadRequest, err))
		return
	}

	if err := tc.TagService.DeleteTag(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Tag deleted successfully", nil)
}

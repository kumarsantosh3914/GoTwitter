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
	page, err := parsePositiveIntQuery(r.URL.Query().Get("page"), "page", 1)
	if err != nil {
		handleError(w, err)
		return
	}

	pageSize, err := parsePositiveIntQuery(r.URL.Query().Get("page_size"), "page_size", 10)
	if err != nil {
		handleError(w, err)
		return
	}

	tags, err := tc.TagService.ListTags(r.Context(), page, pageSize)
	if err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Tags fetched successfully", map[string]any{
		"items": tags,
		"meta": paginationMeta{
			Page:     page,
			PageSize: pageSize,
			Count:    len(tags),
		},
	})
}

func (tc *TagController) GetTagDetails(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handleError(w, apperrors.NewAppError("invalid tag id", http.StatusBadRequest, err))
		return
	}

	page, err := parsePositiveIntQuery(r.URL.Query().Get("page"), "page", 1)
	if err != nil {
		handleError(w, err)
		return
	}

	pageSize, err := parsePositiveIntQuery(r.URL.Query().Get("page_size"), "page_size", 10)
	if err != nil {
		handleError(w, err)
		return
	}

	tag, tweets, err := tc.TagService.GetTagWithTweets(r.Context(), id, page, pageSize)
	if err != nil {
		handleError(w, err)
		return
	}

	response := map[string]interface{}{
		"tag": tag,
		"tweets": map[string]any{
			"items": tweets,
			"meta": paginationMeta{
				Page:     page,
				PageSize: pageSize,
				Count:    len(tweets),
			},
		},
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Tag details fetched successfully", response)
}

func (tc *TagController) GetPopularTags(w http.ResponseWriter, r *http.Request) {
	limit, err := parsePositiveIntQuery(r.URL.Query().Get("limit"), "limit", 10)
	if err != nil {
		handleError(w, err)
		return
	}

	tags, err := tc.TagService.GetPopularTags(r.Context(), limit)
	if err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Popular tags fetched successfully", map[string]any{
		"items": tags,
		"meta": limitMeta{
			Limit: limit,
			Count: len(tags),
		},
	})
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

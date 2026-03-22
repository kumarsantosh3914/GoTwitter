package controllers

import (
	apperrors "GoTwitter/errors"
	"net/http"
	"strconv"
)

type paginationMeta struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Count    int `json:"count"`
}

type limitMeta struct {
	Limit int `json:"limit"`
	Count int `json:"count"`
}

func parsePositiveIntQuery(raw string, field string, defaultValue int) (int, error) {
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, apperrors.NewAppError("invalid "+field, http.StatusBadRequest, nil)
	}

	return value, nil
}

func parsePositiveInt64Query(raw string, field string) (int64, error) {
	if raw == "" {
		return 0, nil
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 1 {
		return 0, apperrors.NewAppError("invalid "+field, http.StatusBadRequest, nil)
	}

	return value, nil
}

package controllers

import (
	apperrors "GoTwitter/errors"
	"GoTwitter/services"
	"GoTwitter/utils"
	"net/http"
)

type MediaController struct {
	MediaService services.MediaService
}

func NewMediaController(mediaService services.MediaService) *MediaController {
	return &MediaController{MediaService: mediaService}
}

func (mc *MediaController) CreatePresignedUpload(w http.ResponseWriter, r *http.Request) {
	claims, ok := utils.GetUserFromContext(r.Context())
	if !ok {
		handleError(w, apperrors.NewAppError("unauthorized", http.StatusUnauthorized, nil))
		return
	}

	var payload struct {
		Filename    string `json:"filename" validate:"required"`
		ContentType string `json:"content_type" validate:"required"`
		SizeBytes   int64  `json:"size_bytes" validate:"required,min=1"`
	}

	if err := utils.ReadJsonBody(r, &payload); err != nil {
		handleError(w, apperrors.NewAppError("invalid json body", http.StatusBadRequest, err))
		return
	}
	if err := utils.Validator.Struct(payload); err != nil {
		handleError(w, apperrors.NewAppError("validation failed", http.StatusBadRequest, err))
		return
	}

	upload, err := mc.MediaService.CreateUpload(r.Context(), claims.UserID, payload.Filename, payload.ContentType, payload.SizeBytes)
	if err != nil {
		handleError(w, err)
		return
	}

	utils.WriteJsonSuccessResponse(w, http.StatusCreated, "Presigned upload created successfully", upload)
}

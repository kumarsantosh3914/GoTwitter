package router

import (
	"GoTwitter/controllers"

	"github.com/go-chi/chi/v5"
)

type MediaRouter struct {
	mediaController *controllers.MediaController
}

func NewMediaRouter(mediaController *controllers.MediaController) Router {
	return &MediaRouter{mediaController: mediaController}
}

func (mr *MediaRouter) Register(r chi.Router) {
	r.Route("/media", func(r chi.Router) {
		r.Use(AuthMiddleware)
		r.Post("/presign", mr.mediaController.CreatePresignedUpload)
	})
}

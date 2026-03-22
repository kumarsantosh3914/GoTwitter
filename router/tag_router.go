package router

import (
	"GoTwitter/controllers"

	"github.com/go-chi/chi/v5"
)

type TagRouter struct {
	tagController *controllers.TagController
}

func NewTagRouter(_tagController *controllers.TagController) Router {
	return &TagRouter{
		tagController: _tagController,
	}
}

func (tr *TagRouter) Register(r chi.Router) {
	r.Route("/tags", func(r chi.Router) {
		r.Get("/", tr.tagController.ListTags)
		r.Get("/popular", tr.tagController.GetPopularTags)
		r.Get("/{id}", tr.tagController.GetTagDetails)

		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware)
			r.Delete("/{id}", tr.tagController.DeleteTag)
		})
	})
}

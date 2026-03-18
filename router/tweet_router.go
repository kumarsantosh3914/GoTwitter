package router

import (
	"GoTwitter/controllers"

	"github.com/go-chi/chi/v5"
)

type TweetRouter struct {
	tweetController *controllers.TweetController
}

func NewTweetRouter(_tweetController *controllers.TweetController) Router {
	return &TweetRouter{
		tweetController: _tweetController,
	}
}

func (tr *TweetRouter) Register(r chi.Router) {
	r.Route("/tweets", func(r chi.Router) {
		r.Get("/", tr.tweetController.ListTweets)
		r.Get("/{id}", tr.tweetController.GetTweet)

		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware)
			r.Post("/", tr.tweetController.CreateTweet)
			r.Put("/{id}", tr.tweetController.UpdateTweet)
			r.Delete("/{id}", tr.tweetController.DeleteTweet)
		})
	})
}

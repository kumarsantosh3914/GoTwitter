package app

import (
	env "GoTwitter/config/env"
	"GoTwitter/controllers"
	db "GoTwitter/db/repositories"
	"GoTwitter/router"
	"GoTwitter/services"
	"log"
	"net/http"
	"strings"
	"time"
)

type Config struct {
	Addr string // PORT
}

type Application struct {
	Config Config
	Store  *db.Storage
}

// Construction for config
func NewConfig() Config {
	port := env.GetString("PORT", ":8080")
	if port != "" && !strings.Contains(port, ":") {
		port = ":" + port
	}

	return Config{
		Addr: port,
	}
}

// Construction for Application
func NewApplication(cfg Config, store *db.Storage) *Application {
	return &Application{
		Config: cfg,
		Store:  store,
	}
}

func (app *Application) Run() error {
	us := services.NewUserService(app.Store.UserRepository)
	uc := controllers.NewUserController(us)
	uRouter := router.NewUserRouter(uc)

	ts := services.NewTweetService(app.Store.TweetRepository)
	tc := controllers.NewTweetController(ts)
	tRouter := router.NewTweetRouter(tc)

	server := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      router.SetupRouter(uRouter, tRouter),
		ReadTimeout:  10 * time.Second, // Set read timeout to 10 seconds
		WriteTimeout: 10 * time.Second, // Set write timeout to 10 seconds
		IdleTimeout:  10 * time.Second, // Set idle timeout to 10 seconds
	}

	log.Println("[INFO] Starting server on", app.Config.Addr)

	return server.ListenAndServe()
}

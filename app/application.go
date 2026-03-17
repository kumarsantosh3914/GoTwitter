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
	Store  db.Storage
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
func NewApplication(cfg Config) *Application {
	return &Application{
		Config: cfg,
		Store:  *db.NewStorage(),
	}
}

func (app *Application) Run() error {
	ur := db.NewUserRepository()
	us := services.NewUserService(ur)
	uc := controllers.NewUserController(us)
	uRouter := router.NewUserRouter(uc)

	server := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      router.SetupRouter(uRouter),
		ReadTimeout:  10 * time.Second, // Set read timeout to 10 seconds
		WriteTimeout: 10 * time.Second, // Set write timeout to 10 seconds
		IdleTimeout:  10 * time.Second, // Set idle timeout to 10 seconds
	}

	log.Println("[INFO] Startring server on", app.Config.Addr)

	return server.ListenAndServe()
}

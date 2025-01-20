package main

import (
	"gin-gorm-api/api"
	"gin-gorm-api/config"
	"gin-gorm-api/middleware"
	"gin-gorm-api/model"
	"gin-gorm-api/provider"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Swagger information

// @title        Gin & Gorm API
// @version      0.1

func logAndExit(err error) {
	log.Fatalf("Failed to start server: %s", err)
}

func startServer(r *gin.Engine) {
	// Create server with timeout
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
		// set timeout due CWE-400 - Potential Slowloris Attack
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		logAndExit(err)
	}
}

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		logAndExit(err)
	}

	db, err := model.NewDBSession(config)
	if err != nil {
		logAndExit(err)
	}
	if err = model.RunMigration(db); err != nil {
		logAndExit(err)
	}

	mailer := provider.NewMailer(config)
	auth, err := provider.NewUserAuthManager(db, mailer, config, "user")
	if err != nil {
		logAndExit(err)
	}
	sm := middleware.NewSessionMiddleware(auth)

	r, err := api.NewEngine(
		config,
		api.NewAuthHandler(auth, sm),
		api.NewUserHandler(db, sm),
	)
	if err != nil {
		logAndExit(err)
	}

	startServer(r)
}

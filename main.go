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

const fatalMessage = "Failed to start server: %s"

func startServer(r *gin.Engine) {
	// Create server with timeout
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
		// set timeout due CWE-400 - Potential Slowloris Attack
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf(fatalMessage, err)
	}
}

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf(fatalMessage, err)
	}

	db, err := model.NewDBSession(config)
	if err != nil {
		log.Fatalf(fatalMessage, err)
	}
	if err = model.RunMigration(db); err != nil {
		log.Fatalf(fatalMessage, err)
	}

	mailer := provider.NewMailer(config)
	auth, err := provider.NewUserAuthManager(db, mailer, config, "user")
	if err != nil {
		log.Fatalf(fatalMessage, err)
	}
	sm := middleware.NewSessionMiddleware(auth)

	r, err := api.NewEngine(config)
	if err != nil {
		log.Fatalf(fatalMessage, err)
	}

	api.NewAuthHandler(auth, sm).AddRoutes(r)
	api.NewUserHandler(db, sm).AddRoutes(r)

	startServer(r)
}

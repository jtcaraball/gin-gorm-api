package main

import (
	"gin-gorm-api/handler"
	"gin-gorm-api/middleware"
	"gin-gorm-api/model"
	"gin-gorm-api/provider"
	"gin-gorm-api/server"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func startServer(r *gin.Engine) {
	// Create server with timeout
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
		// set timeout due CWE-400 - Potential Slowloris Attack
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}

func main() {
	config, err := server.LoadConfig()
	if err != nil {
		panic(err)
	}

	db, err := server.ConnectDB(config)
	if err != nil {
		panic(err)
	}
	if err = model.RunMigration(db); err != nil {
		panic(err)
	}

	mailer := server.NewMailer(config)
	auth, err := provider.NewUserAuthManager(db, config.Secret, "user", mailer)
	if err != nil {
		panic(err)
	}
	sm := middleware.NewSessionMiddleware(auth)

	r, err := server.NewEngine(
		config,
		handler.NewAuthHandler(auth, sm),
		handler.NewUserHandler(db, sm),
	)
	if err != nil {
		panic(err)
	}

	startServer(r)
}

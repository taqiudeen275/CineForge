package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/ats-tech/cineforge/services/control-api/internal/config"
	"github.com/ats-tech/cineforge/services/control-api/internal/httpapi"
	"github.com/ats-tech/cineforge/services/control-api/internal/store"
	"go.temporal.io/sdk/client"
)

func main() {
	ctx := context.Background()
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		log.Error("configuration failed", "error", err)
		os.Exit(1)
	}
	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("database failed", "error", err)
		os.Exit(1)
	}
	defer st.Close()
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: cfg.FirebaseProjectID})
	if err != nil {
		log.Error("firebase failed", "error", err)
		os.Exit(1)
	}
	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Error("firebase auth failed", "error", err)
		os.Exit(1)
	}
	temporalClient, err := client.Dial(client.Options{HostPort: os.Getenv("TEMPORAL_ADDRESS")})
	if err != nil {
		log.Error("temporal failed", "error", err)
		os.Exit(1)
	}
	defer temporalClient.Close()
	h := httpapi.New(cfg, st, authClient, temporalClient, log).Router()
	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: h, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 90 * time.Second}
	go func() {
		log.Info("control api listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server stopped", "error", err)
			os.Exit(1)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdown, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdown)
}

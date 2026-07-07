package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr              string
	Environment           string
	DatabaseURL           string
	WebOrigin             string
	SessionCookieSecure   bool
	FirebaseProjectID     string
	FirebaseAuthEmulator  string
	SessionDurationHours  int
	InvitationDurationDay int
	GCSUploadEndpoint     string
	GCSQuarantineBucket   string
	MailSMTPAddr          string
	MailFrom              string
}

func Load() (Config, error) {
	secure, err := strconv.ParseBool(value("SESSION_COOKIE_SECURE", "true"))
	if err != nil {
		return Config{}, fmt.Errorf("parse SESSION_COOKIE_SECURE: %w", err)
	}
	cfg := Config{
		HTTPAddr:              value("HTTP_ADDR", ":8080"),
		Environment:           value("APP_ENV", "development"),
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		WebOrigin:             value("WEB_ORIGIN", "http://localhost:3000"),
		SessionCookieSecure:   secure,
		FirebaseProjectID:     os.Getenv("FIREBASE_PROJECT_ID"),
		FirebaseAuthEmulator:  os.Getenv("FIREBASE_AUTH_EMULATOR_HOST"),
		SessionDurationHours:  24 * 5,
		InvitationDurationDay: 7,
		GCSUploadEndpoint:     value("GCS_UPLOAD_ENDPOINT", "http://localhost:4443/upload/storage/v1"),
		GCSQuarantineBucket:   value("GCS_BUCKET_QUARANTINE", "cineforge-quarantine"),
		MailSMTPAddr:          value("MAIL_SMTP_ADDR", "localhost:1025"),
		MailFrom:              value("MAIL_FROM", "noreply@cineforge.local"),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.FirebaseProjectID == "" {
		return Config{}, fmt.Errorf("FIREBASE_PROJECT_ID is required")
	}
	return cfg, nil
}

func value(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	// Database file path (default “socialmedia.db”)
	DBPath string

	// Google OAuth credentials and redirect URL
	GoogleClientID      string
	GoogleClientSecret  string
	GoogleRedirectURL   string
	UploadcarePublicKey string
	UploadcareSecretKey string

	// JWT secret key
	JWTSecret string
)

func InitConfig() {
	// Load environment variables from .env (if present)
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	DBPath = os.Getenv("DB_PATH")
	if DBPath == "" {
		DBPath = "socialmedia.db"
	}

	GoogleClientID = os.Getenv("GOOGLE_CLIENT_ID")
	GoogleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	GoogleRedirectURL = os.Getenv("GOOGLE_REDIRECT_URL")
	UploadcarePublicKey = os.Getenv("UPLOADCARE_PUBLIC_KEY")
	UploadcareSecretKey = os.Getenv("UPLOADCARE_SECRET_KEY")

	JWTSecret = os.Getenv("JWT_SECRET")
	if JWTSecret == "" {
		JWTSecret = "secret"
	}
}

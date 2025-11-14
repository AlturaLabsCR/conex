// Package config implements initialization logic for required app parameters
package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Endpoint int

const (
	RootPath Endpoint = iota
	AssetsPath
	EditorPath
	DashboardPath
	RegisterPath
	LoginPath
	LogoutPath
	PricingPath
	AccountPath
	UploadPath
	SettingsPath
)

var Endpoints = map[Endpoint]string{
	RootPath:      "/",
	AssetsPath:    "assets/",
	EditorPath:    "editor/",
	DashboardPath: "dashboard",
	RegisterPath:  "register",
	LoginPath:     "login",
	LogoutPath:    "logout",
	PricingPath:   "pricing",
	AccountPath:   "account",
	UploadPath:    "upload",
	SettingsPath:  "settings",
}

var (
	// Default values are initialized here, these will be used unless overwritten
	// by the Init() method

	AppTitle   string = "CONEX.cr"
	CookieName string = "session"

	// Depends on the RootPrefix, so, must be initialized after checking for any
	// overwrites of RootPrefix

	CSRFHeaderName = "X-CSRF-Token"

	Production bool   = false
	Port       string = "8080"
	LogLevel   int    = 0 // -4:Debug 0:Info 4:Warn 8:Error
	dbDriver   string = "sqlite"
	dbConn     string = "./db.db"

	ServerSMTPUser string
	ServerSMTPHost string
	ServerSMTPPort string
	ServerSMTPPass string

	ServerSecret string
)

const (
	// You should use a prefix for any overwrites via env to avoid conflicts with
	// other programs
	envPrefix = "CONEX_"

	envAppTitle = envPrefix + "TITLE"

	envPort = envPrefix + "PORT"
	envProd = envPrefix + "PROD"
	envLog  = envPrefix + "LOG_LEVEL"
	envDvr  = envPrefix + "DB_DRIVER"
	envCnn  = envPrefix + "DB_CONN"
	envRoot = envPrefix + "ROOT_PREFIX"

	envSMTPUser = envPrefix + "SMTP_USER"
	envSMTPHost = envPrefix + "SMTP_HOST"
	envSMTPPort = envPrefix + "SMTP_PORT"
	envSMTPPass = envPrefix + "SMTP_PASS"

	envCookieName   = envPrefix + "COOKIE_NAME"
	envServerSecret = envPrefix + "SECRET"
)

func Init() {
	godotenv.Load()

	a := os.Getenv(envAppTitle)
	if a != "" {
		AppTitle = a
	}

	r := os.Getenv(envRoot)
	if r != "" {
		Endpoints[RootPath] = r
	}

	c := os.Getenv(envCookieName)
	if c != "" {
		CookieName = c
	}

	// Prefix all endpoint paths with Endpoints[RootPath]
	for key, path := range Endpoints {
		if key == RootPath {
			continue
		}
		Endpoints[key] = Endpoints[RootPath] + path
	}

	Production = os.Getenv(envProd) == "1"

	p := os.Getenv(envPort)
	if p != "" {
		Port = p
	}

	ServerSMTPUser = os.Getenv(envSMTPUser)
	ServerSMTPHost = os.Getenv(envSMTPHost)
	ServerSMTPPort = os.Getenv(envSMTPPort)
	ServerSMTPPass = os.Getenv(envSMTPPass)

	if ServerSMTPUser == "" || ServerSMTPHost == "" || ServerSMTPPort == "" || ServerSMTPPass == "" {
		panic("Required SMTP credentials are not set")
	}

	ServerSecret = os.Getenv(envServerSecret)

	if ServerSecret == "" {
		ServerSecret = generateRandomSecret(32)
	}

	logLevelStr := os.Getenv(envLog)
	if logLevelStr != "" {
		var err error
		l, err := strconv.Atoi(logLevelStr)
		if err == nil {
			LogLevel = l
		}
	}

	dvr := os.Getenv(envDvr)
	if dvr != "" {
		dbDriver = dvr
	}

	conn := os.Getenv(envCnn)
	if conn != "" {
		dbConn = conn
	}
}

func generateRandomSecret(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// never return predictable bytes
		panic(fmt.Errorf("failed to generate random secret: %w", err))
	}
	return hex.EncodeToString(b)
}

// Package config implements initialization logic for required app parameters
package config

import (
	"os"
	"strconv"
)

var (
	// Default values are initialized here, these will be used unless overwritten
	// by the Init() method

	AppTitle string = "CONEX.cr"

	RootPrefix string = "/"
	EditorPath string = RootPrefix + "e"
	// Depends on the RootPrefix, so, must be initialized after checking for any
	// overwrites of RootPrefix
	Assets string

	Production bool   = false
	Port       string = "8080"
	LogLevel   int    = 0 // -4:Debug 0:Info 4:Warn 8:Error
	dbDriver   string = "sqlite"
	dbConn     string = "./db.db"
)

const (
	// You should use a prefix for any overwrites via env to avoid conflicts with
	// other programs
	envPrefix = "APP_"
	envPort   = envPrefix + "PORT"
	envProd   = envPrefix + "PROD"
	envLog    = envPrefix + "LOG_LEVEL"
	envDvr    = envPrefix + "DB_DRIVER"
	envCnn    = envPrefix + "DB_CONN"
	envRoot   = envPrefix + "ROOT_PREFIX"
)

func Init() {
	r := os.Getenv(envRoot)
	if r != "" {
		RootPrefix = r
	}
	Assets = RootPrefix + "assets/"

	Production = os.Getenv(envProd) == "1"

	p := os.Getenv(envPort)
	if p != "" {
		Port = p
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

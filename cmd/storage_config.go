package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/viper"
	storage "github.com/tavocg/go-storage"
	fsbackend "github.com/tavocg/go-storage/backends/fs"
	s3backend "github.com/tavocg/go-storage/backends/s3"
)

const (
	defaultStorageBackend = "fs"
	defaultStorageRoot    = "./data/root"
	defaultStorageMaxSize = int64(10 * 1024 * 1024 * 1024)
	defaultS3Region       = "auto"
)

func initStorageConfig() {
	viper.SetDefault("storage.backend", defaultStorageBackend)
	viper.SetDefault("storage.fs.root", defaultStorageRoot)
	viper.SetDefault("storage.fs.max-size", defaultStorageMaxSize)
	viper.SetDefault("storage.s3.bucket", "")
	viper.SetDefault("storage.s3.access-key-id", "")
	viper.SetDefault("storage.s3.secret-access-key", "")
	viper.SetDefault("storage.s3.region", defaultS3Region)
	viper.SetDefault("storage.s3.endpoint", "")
	viper.SetDefault("storage.s3.max-size", defaultStorageMaxSize)

	flags := rootCmd.PersistentFlags()
	flags.String("storage-backend", defaultStorageBackend, "storage backend to use")
	flags.String("storage-fs-root", defaultStorageRoot, "filesystem storage root directory")
	flags.Int64("storage-fs-max-size", defaultStorageMaxSize, "maximum total stored bytes for filesystem storage")
	flags.String("storage-s3-bucket", "", "S3 bucket name")
	flags.String("storage-s3-access-key-id", "", "S3 access key ID")
	flags.String("storage-s3-secret-access-key", "", "S3 secret access key")
	flags.String("storage-s3-region", defaultS3Region, "S3 region")
	flags.String("storage-s3-endpoint", "", "S3-compatible endpoint URL")
	flags.Int64("storage-s3-max-size", defaultStorageMaxSize, "maximum total stored bytes for S3 storage")

	mustBindPersistentFlag("storage.backend", rootCmd, "storage-backend")
	mustBindPersistentFlag("storage.fs.root", rootCmd, "storage-fs-root")
	mustBindPersistentFlag("storage.fs.max-size", rootCmd, "storage-fs-max-size")
	mustBindPersistentFlag("storage.s3.bucket", rootCmd, "storage-s3-bucket")
	mustBindPersistentFlag("storage.s3.access-key-id", rootCmd, "storage-s3-access-key-id")
	mustBindPersistentFlag("storage.s3.secret-access-key", rootCmd, "storage-s3-secret-access-key")
	mustBindPersistentFlag("storage.s3.region", rootCmd, "storage-s3-region")
	mustBindPersistentFlag("storage.s3.endpoint", rootCmd, "storage-s3-endpoint")
	mustBindPersistentFlag("storage.s3.max-size", rootCmd, "storage-s3-max-size")

	mustBindEnv("storage.backend", "CONEX_STORAGE_BACKEND")
	mustBindEnv("storage.fs.root", "CONEX_FS_ROOT")
	mustBindEnv("storage.fs.max-size", "CONEX_FS_MAX_SIZE")
	mustBindEnv("storage.s3.bucket", "CONEX_S3_BUCKET")
	mustBindEnv("storage.s3.access-key-id", "CONEX_S3_ACCESS_KEY_ID")
	mustBindEnv("storage.s3.secret-access-key", "CONEX_S3_SECRET_ACCESS_KEY")
	mustBindEnv("storage.s3.region", "CONEX_S3_REGION")
	mustBindEnv("storage.s3.endpoint", "CONEX_S3_ENDPOINT")
	mustBindEnv("storage.s3.max-size", "CONEX_S3_MAX_SIZE")
}

func newStorageFromConfig(ctx context.Context) (*storage.Storage, error) {
	switch backend := strings.ToLower(strings.TrimSpace(viper.GetString("storage.backend"))); backend {
	case "fs", "":
		return fsbackend.New(
			ctx,
			fsbackend.WithRoot(viper.GetString("storage.fs.root")),
			fsbackend.WithMaxSize(viper.GetInt64("storage.fs.max-size")),
		)
	case "s3":
		return s3backend.New(
			ctx,
			s3backend.WithBucket(viper.GetString("storage.s3.bucket")),
			s3backend.WithAccessKey(viper.GetString("storage.s3.access-key-id")),
			s3backend.WithSecretKey(viper.GetString("storage.s3.secret-access-key")),
			s3backend.WithRegion(viper.GetString("storage.s3.region")),
			s3backend.WithEndpoint(viper.GetString("storage.s3.endpoint")),
			s3backend.WithMaxSize(viper.GetInt64("storage.s3.max-size")),
		)
	default:
		return nil, fmt.Errorf("unsupported storage backend %q", backend)
	}
}

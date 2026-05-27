package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"github.com/tavocg/go-storage"
	"github.com/tavocg/go-storage/backends/fs"
	"github.com/tavocg/go-storage/backends/s3"
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
	viper.SetDefault("storage.s3.private-bucket", "")
	viper.SetDefault("storage.s3.public-bucket", "")
	viper.SetDefault("storage.s3.access-key-id", "")
	viper.SetDefault("storage.s3.secret-access-key", "")
	viper.SetDefault("storage.s3.region", defaultS3Region)
	viper.SetDefault("storage.s3.endpoint", "")
	viper.SetDefault("storage.s3.public-endpoint-url", "")
	viper.SetDefault("storage.s3.max-size", defaultStorageMaxSize)

	flags := rootCmd.PersistentFlags()
	flags.String("storage-backend", defaultStorageBackend, "storage backend to use")
	flags.String("storage-fs-root", defaultStorageRoot, "filesystem storage root directory")
	flags.Int64("storage-fs-max-size", defaultStorageMaxSize, "maximum total stored bytes for filesystem storage")
	flags.String("storage-s3-private-bucket", "", "private S3 bucket name")
	flags.String("storage-s3-public-bucket", "", "public S3 bucket name")
	flags.String("storage-s3-access-key-id", "", "S3 access key ID")
	flags.String("storage-s3-secret-access-key", "", "S3 secret access key")
	flags.String("storage-s3-region", defaultS3Region, "S3 region")
	flags.String("storage-s3-endpoint", "", "S3-compatible endpoint URL")
	flags.String("storage-s3-public-endpoint-url", "", "public S3 endpoint URL")
	flags.Int64("storage-s3-max-size", defaultStorageMaxSize, "maximum total stored bytes for S3 storage")

	mustBindPersistentFlag("storage.backend", rootCmd, "storage-backend")
	mustBindPersistentFlag("storage.fs.root", rootCmd, "storage-fs-root")
	mustBindPersistentFlag("storage.fs.max-size", rootCmd, "storage-fs-max-size")
	mustBindPersistentFlag("storage.s3.private-bucket", rootCmd, "storage-s3-private-bucket")
	mustBindPersistentFlag("storage.s3.public-bucket", rootCmd, "storage-s3-public-bucket")
	mustBindPersistentFlag("storage.s3.access-key-id", rootCmd, "storage-s3-access-key-id")
	mustBindPersistentFlag("storage.s3.secret-access-key", rootCmd, "storage-s3-secret-access-key")
	mustBindPersistentFlag("storage.s3.region", rootCmd, "storage-s3-region")
	mustBindPersistentFlag("storage.s3.endpoint", rootCmd, "storage-s3-endpoint")
	mustBindPersistentFlag("storage.s3.public-endpoint-url", rootCmd, "storage-s3-public-endpoint-url")
	mustBindPersistentFlag("storage.s3.max-size", rootCmd, "storage-s3-max-size")

	mustBindEnv("storage.backend", "CONEX_STORAGE_BACKEND")
	mustBindEnv("storage.fs.root", "CONEX_FS_ROOT")
	mustBindEnv("storage.fs.max-size", "CONEX_FS_MAX_SIZE")
	mustBindEnv("storage.s3.private-bucket", "CONEX_S3_PRIVATE_BUCKET")
	mustBindEnv("storage.s3.public-bucket", "CONEX_S3_PUBLIC_BUCKET")
	mustBindEnv("storage.s3.access-key-id", "CONEX_S3_ACCESS_KEY_ID")
	mustBindEnv("storage.s3.secret-access-key", "CONEX_S3_SECRET_ACCESS_KEY")
	mustBindEnv("storage.s3.region", "CONEX_S3_REGION")
	mustBindEnv("storage.s3.endpoint", "CONEX_S3_ENDPOINT")
	mustBindEnv("storage.s3.public-endpoint-url", "CONEX_S3_PUBLIC_ENDPOINT_URL")
	mustBindEnv("storage.s3.max-size", "CONEX_S3_MAX_SIZE")
}

func newStorageFromConfig(ctx context.Context) (*storage.Storage, *storage.Storage, error) {
	switch backend := strings.ToLower(strings.TrimSpace(viper.GetString("storage.backend"))); backend {
	case "fs", "":
		privateStorage, err := fs.New(
			ctx,
			fs.WithRoot(filepath.Join(viper.GetString("storage.fs.root"), "private")),
			fs.WithMaxSize(viper.GetInt64("storage.fs.max-size")),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("create private fs storage: %w", err)
		}

		publicStorage, err := fs.New(
			ctx,
			fs.WithRoot(filepath.Join(viper.GetString("storage.fs.root"), "public")),
			fs.WithMaxSize(viper.GetInt64("storage.fs.max-size")),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("create public fs storage: %w", err)
		}

		return privateStorage, publicStorage, nil
	case "s3":
		privateBucket := strings.TrimSpace(viper.GetString("storage.s3.private-bucket"))
		publicBucket := strings.TrimSpace(viper.GetString("storage.s3.public-bucket"))
		if privateBucket == "" {
			return nil, nil, fmt.Errorf("private S3 bucket is required")
		}
		if publicBucket == "" {
			return nil, nil, fmt.Errorf("public S3 bucket is required")
		}
		if privateBucket == publicBucket {
			return nil, nil, fmt.Errorf("private and public S3 buckets must be different")
		}

		privateStorage, err := s3.New(
			ctx,
			s3.WithBucket(privateBucket),
			s3.WithAccessKey(viper.GetString("storage.s3.access-key-id")),
			s3.WithSecretKey(viper.GetString("storage.s3.secret-access-key")),
			s3.WithRegion(viper.GetString("storage.s3.region")),
			s3.WithEndpoint(viper.GetString("storage.s3.endpoint")),
			s3.WithMaxSize(viper.GetInt64("storage.s3.max-size")),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("create private S3 storage: %w", err)
		}

		publicStorage, err := s3.New(
			ctx,
			s3.WithBucket(publicBucket),
			s3.WithAccessKey(viper.GetString("storage.s3.access-key-id")),
			s3.WithSecretKey(viper.GetString("storage.s3.secret-access-key")),
			s3.WithRegion(viper.GetString("storage.s3.region")),
			s3.WithEndpoint(viper.GetString("storage.s3.endpoint")),
			s3.WithPublicEndpointURL(viper.GetString("storage.s3.public-endpoint-url")),
			s3.WithMaxSize(viper.GetInt64("storage.s3.max-size")),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("create public S3 storage: %w", err)
		}

		return privateStorage, publicStorage, nil
	default:
		return nil, nil, fmt.Errorf("unsupported storage backend %q", backend)
	}
}

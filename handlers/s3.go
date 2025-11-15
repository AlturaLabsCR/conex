package handlers

import (
	"bytes"
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"time"

	"app/config"
	"app/internal/db"
	"app/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (h *Handler) PutObject(ctx context.Context, siteSlug, fileName string, body io.Reader) (string, error) {
	mime, size, data, err := utils.InspectReader(body)
	if err != nil {
		return "", err
	}

	md5sum := md5.Sum(data)
	md5 := hex.EncodeToString(md5sum[:])

	tx, err := h.DB().Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	queries := db.New(tx)

	site, err := queries.GetSiteBySlug(ctx, siteSlug)
	if err != nil {
		return "", err
	}

	objKey := site.SiteSlug + "/" + fileName
	objURL := config.S3PublicURL + "/" + objKey

	now := time.Now().Unix()

	var obj db.SiteObject

	obj, err = queries.GetObject(ctx, db.GetObjectParams{
		ObjectBucket: config.S3Bucket,
		ObjectKey:    objKey,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			obj, err = queries.InsertObject(ctx, db.InsertObjectParams{
				ObjectSite:         site.SiteID,
				ObjectBucket:       config.S3Bucket,
				ObjectKey:          objKey,
				ObjectMime:         mime,
				ObjectMd5:          md5,
				ObjectSizeBytes:    size,
				ObjectCreatedUnix:  now,
				ObjectModifiedUnix: now,
			})
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	} else {
		if md5 == obj.ObjectMd5 {
			return objURL, nil
		} else {
			if err := queries.UpdateObject(ctx, db.UpdateObjectParams{
				ObjectID:           obj.ObjectID,
				ObjectMime:         mime,
				ObjectMd5:          md5,
				ObjectSizeBytes:    size,
				ObjectModifiedUnix: now,
			}); err != nil {
				return "", err
			}
			// obj.ObjectMime = mime
			// obj.ObjectMd5 = md5
			// obj.ObjectSizeBytes = size
			// obj.ObjectModifiedUnix = now
		}
	}

	if _, err := h.S3().PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(config.S3Bucket),
		Key:    aws.String(objKey),
		Body:   bytes.NewReader(data),
	}); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return objURL, nil
}

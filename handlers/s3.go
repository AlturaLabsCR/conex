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
	"app/database"
	"app/internal/db"
	"app/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (h *Handler) PutObject(ctx context.Context, siteSlug, fileName string, body io.Reader, queries *db.Queries) (db.SiteObject, error) {
	mime, size, data, err := utils.InspectReader(body)
	if err != nil {
		return db.SiteObject{}, err
	}

	md5sum := md5.Sum(data)
	md5 := hex.EncodeToString(md5sum[:])

	site, err := queries.GetSiteBySlug(ctx, siteSlug)
	if err != nil {
		return db.SiteObject{}, err
	}

	objKey := site.SiteSlug + "/" + fileName

	if err := database.ValidateObjectStrings(config.S3Bucket, objKey, mime, md5); err != nil {
		return db.SiteObject{}, err
	}

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
				return db.SiteObject{}, err
			}
		} else {
			return db.SiteObject{}, err
		}
	} else {
		if md5 == obj.ObjectMd5 {
			return obj, nil
		} else {
			if err := queries.UpdateObject(ctx, db.UpdateObjectParams{
				ObjectID:           obj.ObjectID,
				ObjectMime:         mime,
				ObjectMd5:          md5,
				ObjectSizeBytes:    size,
				ObjectModifiedUnix: now,
			}); err != nil {
				return db.SiteObject{}, err
			}

			obj.ObjectMime = mime
			obj.ObjectMd5 = md5
			obj.ObjectSizeBytes = size
			obj.ObjectModifiedUnix = now
		}
	}

	if _, err := h.S3().PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(config.S3Bucket),
		Key:    aws.String(objKey),
		Body:   bytes.NewReader(data),
	}); err != nil {
		return db.SiteObject{}, err
	}

	return obj, nil
}

func (h *Handler) DeleteObject(ctx context.Context, key string, queries *db.Queries) error {
	if _, err := h.S3().DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(config.S3Bucket),
		Key:    aws.String(key),
	}); err != nil {
		return err
	}

	err := queries.DeleteObject(ctx, db.DeleteObjectParams{
		ObjectKey:    key,
		ObjectBucket: config.S3Bucket,
	})

	return err
}

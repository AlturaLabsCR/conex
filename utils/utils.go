// Package utils implements helper methods for common misc tasks
package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func RandomString() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func Hash(s string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(s), 10)
	return string(bytes), err
}

func HashCompare(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

func HostURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

func InspectReader(r io.Reader) (mime string, size int64, data []byte, err error) {
	data, err = io.ReadAll(r)
	if err != nil {
		return "", 0, nil, err
	}

	size = int64(len(data))

	mime = http.DetectContentType(data)

	return mime, size, data, nil
}

func MD5FromReader(r io.Reader) (string, error) {
	h := md5.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	sum := h.Sum(nil)
	return hex.EncodeToString(sum), nil
}

func UnixToYMD(timestamp int64) string {
	loc, _ := time.LoadLocation("America/Costa_Rica")
	t := time.Unix(timestamp, 0).In(loc)
	return t.Format("2006-01-02")
}

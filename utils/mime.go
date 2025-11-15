package utils

import (
	"io"
	"net/http"
)

func InspectReader(r io.Reader) (mime string, size int64, data []byte, err error) {
	data, err = io.ReadAll(r)
	if err != nil {
		return "", 0, nil, err
	}

	size = int64(len(data))

	mime = http.DetectContentType(data)

	return mime, size, data, nil
}

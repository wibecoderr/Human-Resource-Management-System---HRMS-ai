package utils

import (
	"encoding/json"
	"io"
)

func ParseBody(body io.ReadCloser, dst interface{}) error {
	defer body.Close()
	return json.NewDecoder(body).Decode(dst)
}

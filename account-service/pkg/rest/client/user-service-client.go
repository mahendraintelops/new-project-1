package client

import (
	"io"
	"net/http"
)

func ExecuteNodeEe() ([]byte, error) {
	response, err := http.Get("user-service:3000/ping")

	if err != nil {
		return nil, err
	}

	return io.ReadAll(response.Body)
}

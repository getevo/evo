package network

import (
	"net/http"
)

// HttpStatusCode return http status code of url
func HttpStatusCode(url string) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return -1, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}

package proxy

import (
	"compress/gzip"
	"io"
	"net/http"
)

type ParsedResponse struct {
	StatusCode    int
	Status        string
	Headers       map[string][]string
	Body          string
	ContentLength int64
	Compressed    bool
}

func ParseResponse(resp *http.Response) (*ParsedResponse, error) {
	parsedResp := &ParsedResponse{
		StatusCode:    resp.StatusCode,
		Status:        resp.Status,
		Headers:       make(map[string][]string),
		ContentLength: resp.ContentLength,
	}

	for name, values := range resp.Header {
		parsedResp.Headers[name] = values
	}

	var bodyReader io.ReadCloser = resp.Body

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		parsedResp.Compressed = true
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		bodyReader = reader
	}

	bodyBytes, err := io.ReadAll(bodyReader)
	if err != nil {
		return nil, err
	}
	parsedResp.Body = string(bodyBytes)

	return parsedResp, nil
}

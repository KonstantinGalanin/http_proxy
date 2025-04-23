package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

type ParsedRequest struct {
	Scheme     string              `json:"scheme"`
	Method     string              `json:"method"`
	Path       string              `json:"path"`
	Host       string              `json:"host"`
	Headers    map[string][]string `json:"headers"`
	Cookies    map[string]string   `json:"cookies"`
	GetParams  map[string][]string `json:"get_params"`
	PostParams map[string][]string `json:"post_params"`
	Body       string              `json:"body"`
}

func ParseRequest(r *http.Request, scheme string) (*ParsedRequest, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	cookies := map[string]string{}
	for _, cookie := range r.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}

	req := &ParsedRequest{
		Scheme:     scheme,
		Method:     r.Method,
		Path:       r.URL.Path,
		Host:       r.Host,
		Headers:    r.Header,
		Cookies:    cookies,
		GetParams:  r.URL.Query(),
		PostParams: r.PostForm,
		Body:       string(bodyBytes),
	}

	return req, nil
}

func BuildRequest(parsed *ParsedRequest) (*http.Request, error) {
	u := &url.URL{
		Scheme: parsed.Scheme,
		Path:   parsed.Path,
		Host:   parsed.Host,
	}

	q := u.Query()
	for key, values := range parsed.GetParams {
		for _, value := range values {
			q.Add(key, value)
		}
	}
	u.RawQuery = q.Encode()

	var body io.Reader
	if parsed.Body != "" {
		body = bytes.NewBufferString(parsed.Body)
	}

	req, err := http.NewRequest(parsed.Method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Host = parsed.Host

	for key, values := range parsed.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	for name, value := range parsed.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}

	if (parsed.Method == http.MethodPost || parsed.Method == http.MethodPut) && len(parsed.PostParams) > 0 {
		form := make(url.Values)
		for key, values := range parsed.PostParams {
			for _, value := range values {
				form.Add(key, value)
			}
		}
		req.Body = io.NopCloser(bytes.NewBufferString(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return req, nil
}

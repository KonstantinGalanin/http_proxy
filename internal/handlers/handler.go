package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/KonstantinGalanin/http_proxy/internal/proxy"
	"github.com/gorilla/mux"
)

var (
	ErrInvalidID = errors.New("invalid id")
)

type Repository interface {
	GetRequests() ([]*proxy.ParsedRequest, error)
	GetRequestByID(id int) (*proxy.ParsedRequest, error)
}

type Handler struct {
	repo Repository
}

func New(repo Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

func (h *Handler) GetRequests(w http.ResponseWriter, r *http.Request) {
	fmt.Println("get request")
	requests, err := h.repo.GetRequests()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(requests); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetRequestByID(w http.ResponseWriter, r *http.Request) {
	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, ErrInvalidID.Error(), http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	request, err := h.repo.GetRequestByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(request); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) RepeatRequest(w http.ResponseWriter, r *http.Request) {
	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, ErrInvalidID.Error(), http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	parsedRequest, err := h.repo.GetRequestByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	request, err := proxy.BuildRequest(parsedRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.Copy(w, response.Body)
	response.Body.Close()
}

func sendInjectedRequest(w http.ResponseWriter, req *proxy.ParsedRequest, response *http.Response) (bool, error) {
	injectedRequest, err := proxy.BuildRequest(req)
		if err != nil {
			return false, err
		}
	injectedResponse, err := http.DefaultTransport.RoundTrip(injectedRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false, err
	}

	if injectedResponse.StatusCode != response.StatusCode {
		return true, nil
	}

	return false, nil
}

func (h *Handler) ScanRequest(w http.ResponseWriter, r *http.Request) {
	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, ErrInvalidID.Error(), http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	parsedRequest, err := h.repo.GetRequestByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	request, err := proxy.BuildRequest(parsedRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	injections := []string{`'`, `"`}
	for _, injection := range injections {
		pathInjReq := parsedRequest
		pathInjReq.Path += injection

		cookiesInjReq := parsedRequest
		for cookieKey := range cookiesInjReq.Cookies{
			cookiesInjReq.Cookies[cookieKey] += injection
		}

		headersInjReq := parsedRequest
		for headerKey := range headersInjReq.Headers{
			if len(headersInjReq.Headers[headerKey]) > 0 {
				lenHeaderVals := len(headersInjReq.Headers[headerKey])
				headersInjReq.Headers[headerKey][lenHeaderVals - 1] += injection
			}
		}

		injectedRequests := []*proxy.ParsedRequest{pathInjReq, cookiesInjReq, headersInjReq}
		for _, injectedReq := range injectedRequests {
			isInjected, err := sendInjectedRequest(w, injectedReq, response)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			if isInjected {
				w.Write([]byte("Possible SQL Injection"))
				return
			}
		}
	}

	w.Write([]byte("No SQL Injections"))
}

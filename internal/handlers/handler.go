package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	ErrInvalidID = errors.New("invalid id")
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) GetRequests(w http.ResponseWriter, r *http.Request) {
	fmt.Println("test1")
}

func (h *Handler) GetRequest(w http.ResponseWriter, r *http.Request) {
	_, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, ErrInvalidID.Error(), http.StatusBadRequest)
		return
	}

}

func (h *Handler) RepeatRequest(w http.ResponseWriter, r *http.Request) {
	_, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, ErrInvalidID.Error(), http.StatusBadRequest)
		return
	}

}

func (h *Handler) ScanRequest(w http.ResponseWriter, r *http.Request) {
	_, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, ErrInvalidID.Error(), http.StatusBadRequest)
		return
	}

}

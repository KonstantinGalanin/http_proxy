package main

import (
	"net/http"

	"github.com/KonstantinGalanin/http_proxy/internal/handlers"
	"github.com/gorilla/mux"
)

func main() {
	handler := handlers.NewHandler()

	r := mux.NewRouter()

	r.HandleFunc("/requests", handler.GetRequests).Methods(http.MethodGet)
	r.HandleFunc("/requests/{id}", handler.GetRequest).Methods(http.MethodGet)
	r.HandleFunc("/repeat/{id}", handler.RepeatRequest).Methods(http.MethodGet)
	r.HandleFunc("/scan/{id}", handler.ScanRequest).Methods(http.MethodGet)

	http.ListenAndServe(":8000", r)
}

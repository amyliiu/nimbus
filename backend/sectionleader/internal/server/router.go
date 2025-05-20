package server

import (
	"net/http"

	"github.com/tongshengw/nimbus/backend/sectionleader/internal/handler"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", handler.HelloHandler)
	return mux
}

package main

import (
	"log"
	"net/http"

	"github.com/tongshengw/nimbus/backend/sectionleader/internal/server"
)

func main() {
	r := server.NewRouter()

	addr := ":8080"
	log.Printf("Starting server at http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Println(err)
	}
}

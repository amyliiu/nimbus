package handler

import (
	"net/http"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
)

type requestData struct {
}

func NewVMHandler(w http.ResponseWriter, r *http.Request) {
	firecracker.
}
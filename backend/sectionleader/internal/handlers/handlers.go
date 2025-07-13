package handlers

import (
	"encoding/json"
	"net/http"
)

func CheckStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write(jsonBytes)
}

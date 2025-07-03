package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/middle"
)

func NewMachine(w http.ResponseWriter, r *http.Request) {
	data, ok := r.Context().Value(middle.CommonContextDataKey).(middle.CommonContextData)
	if !ok {
		logrus.Errorf("common context data not ok: %v", data)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	vmManager := data.Manager

	id, err := vmManager.CreateVM()
	if err != nil {
		logrus.Errorf("create vm failed: %v", err)
		http.Error(w, "Failed to create VM", http.StatusInternalServerError)
		return
	}
	
	tokenStr, err := middle.NewJwt(id, data.SecretKey)
	if err != nil {
		logrus.Errorf("new jwt failed: %v", err)
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}

	response := struct {
		MachineId string `json:"machine-id"`
		Token     string `json:"token"`
	}{
		MachineId: id.String(),
		Token:    tokenStr,
	}

	w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

func StopMachine(w http.ResponseWriter, r *http.Request) {

}
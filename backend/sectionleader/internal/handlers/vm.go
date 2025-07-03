package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/app"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/constants"
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

	outputChan, err := vmManager.CreateVM()
	if err != nil {
		logrus.Errorf("create vm failed: %v", err)
		http.Error(w, "Failed to create VM", http.StatusInternalServerError)
		return
	}

	var createMachineRes *app.MachineData

	select {
	case createMachineRes = <-outputChan:
		if createMachineRes == nil {
			logrus.Errorf("create vm returned nil data pointer")
			http.Error(w, "Failed to create VM", http.StatusInternalServerError)
			return
		}
	case <-time.After(constants.CreateVmTimeout):
		logrus.Errorf("create machine timed out")
		http.Error(w, "timed out creating VM", http.StatusInternalServerError)
		return
	}

	tokenStr, err := middle.NewJwt(createMachineRes.Id, data.SecretKey)
	if err != nil {
		logrus.Errorf("new jwt failed: %v", err)
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}

	response := struct {
		MachineId string `json:"machine_id"`
		MachineName string `json:"machine_name"`
		Token     string `json:"token"`
	}{
		MachineId: createMachineRes.Id.String(),
		MachineName: createMachineRes.Name,
		Token:     tokenStr,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func StopMachine(w http.ResponseWriter, r *http.Request) {

}

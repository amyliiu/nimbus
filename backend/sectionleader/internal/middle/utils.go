package middle

import (
	"context"
	"net/http"

	"github.com/tongshengw/nimbus/backend/sectionleader/internal/app"
)

type ContextKey string
const CommonContextDataKey ContextKey = "request-data"
const MachineIdContextDataKey ContextKey = "user-machine-id"

type CommonContextData struct {
	Manager *app.VMManager
	SecretKey string
}

func WithData(data CommonContextData, next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), CommonContextDataKey, data)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
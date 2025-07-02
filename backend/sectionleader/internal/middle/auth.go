package middle

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/app"
)

func NewJwt(id app.MachineUUID) {
	
}

func CheckJwt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.Infof("Jwt auth review")
		next.ServeHTTP(w, r)
	})
}
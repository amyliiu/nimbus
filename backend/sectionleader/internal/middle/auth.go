package middle

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func JwtAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.Infof("HEREHEREHERE")
	})
}
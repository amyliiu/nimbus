package middle

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.Infof("Request recieved. Method: %s, Path: %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
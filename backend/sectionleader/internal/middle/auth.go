package middle

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/app"
)

func NewJwt(id app.MachineUUID, secretKey string) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"machineId": id.String(),
			"iat": time.Now().Unix(),
		})

	tokenStr, err := token.SignedString(secretKey)
	if err != nil {
		logrus.Errorf("new jwt signedstring failed: %v", err)
		return "", err
	}

	return tokenStr, nil
}

func CheckJwt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.Infof("Jwt auth review")
		next.ServeHTTP(w, r)
	})
}
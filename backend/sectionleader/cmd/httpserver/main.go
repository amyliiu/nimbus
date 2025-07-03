package main

import (
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/app"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/middle"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/handlers"
	"github.com/joho/godotenv"
)

func main() {
	logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logrus.Fatalf("failed to open log file: %v", err)
	}
	logrus.SetOutput(logFile)
	
	err = godotenv.Load()
	if err != nil {
		logrus.Fatalf("failed to load .env: %v", err)
	}
	
	secretKey := os.Getenv("SECRET_KEY")

	vmManager := app.NewVMManager()
	app.InstallSignalHandlers(vmManager)

	mux := http.NewServeMux()
	mux.Handle("POST /new-machine", http.HandlerFunc(handlers.NewMachine))
	
	privateMux := http.NewServeMux()
    privateMux.Handle("POST /stop-machine", http.HandlerFunc(handlers.StopMachine))

	mux.Handle("/private/", http.StripPrefix("/private", middle.CheckJwt(privateMux)))
	
    logrus.Println("Starting server on :8080")
	
	commonContextData := middle.CommonContextData{
		Manager: vmManager,
		SecretKey: secretKey,
	}
	
	server := http.Server{
		Addr: ":8080"	,
		Handler: middle.LogRequest(middle.WithData(commonContextData, mux)),
	}
	err = server.ListenAndServe()
    if err != nil {
        logrus.Fatalf("server failed: %v", err)
    }
}
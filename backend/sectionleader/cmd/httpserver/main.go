package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/app"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/handlers"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/middle"
)

func main() {
	// Clear some default handlers installed by the firecracker SDK:
	signal.Reset(os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logrus.Fatalf("failed to open log file: %v", err)
	}
	logrus.SetOutput(logFile)

	err = godotenv.Load()
	if err != nil {
		logrus.Fatalf("failed to load .env: %v", err)
	}

	vmManager := app.NewVMManager()
	installSignalHandlers(vmManager)

	mux := http.NewServeMux()
	mux.Handle("POST /new-machine", http.HandlerFunc(handlers.NewMachine))
	mux.Handle("POST /shutdown-all", http.HandlerFunc(handlers.ShutdownAll))

	privateMux := http.NewServeMux()
	privateMux.Handle("GET /ssh-key", http.HandlerFunc(handlers.SshKey))
	privateMux.Handle("POST /stop-machine", http.HandlerFunc(handlers.StopMachine))

	mux.Handle("/private/", http.StripPrefix("/private", middle.CheckJwt(privateMux)))

	commonContextData := middle.CommonContextData{
		Manager:   vmManager,
		SecretKey: os.Getenv("SECRET_KEY"),
	}

	splash := `
 ____  ____  ___  ____  __  __   __ _  __    ____   __   ____  ____  ____ 
/ ___)(  __)/ __)(_  _)(  )/  \ (  ( \(  )  (  __) / _\ (    \(  __)(  _ \
\___ \ ) _)( (__   )(   )((  O )/    // (_/\ ) _) /    \ ) D ( ) _)  )   /
(____/(____)\___) (__) (__)\__/ \_)__)\____/(____)\_/\_/(____/(____)(__\_)

`
	fmt.Print(splash)
	logrus.Println("Starting server on :8080")
	fmt.Println("Starting server on :8080")

	server := http.Server{
		Addr:    ":8080",
		Handler: middle.LogRequest(middle.WithData(commonContextData, mux)),
	}
	err = server.ListenAndServe()
	if err != nil {
		logrus.Fatalf("server failed: %v", err)
	}

}

func installSignalHandlers(manager *app.VMManager) {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

		for {
			switch s := <-c; {
			case s == syscall.SIGTERM || s == os.Interrupt:
				logrus.Printf("Caught signal: %s, requesting clean shutdown", s.String())
				err := manager.GracefulShutdownAll()
				if err != nil {
					logrus.Errorf("An error occurred while stopping Firecracker VMM: %v", err)
				}
				logrus.Infof("exiting from sigterm or interrupt")
				fmt.Println()
				os.Exit(0)
			case s == syscall.SIGQUIT:
				// FIXME: force shutdown
				logrus.Printf("Caught signal: %s, forcing shutdown", s.String())
				// if err := m.StopVMM(); err != nil {
				// 	logrus.Errorf("An error occurred while stopping Firecracker VMM: %v", err)
				// }
				err := manager.GracefulShutdownAll()
				if err != nil {
					logrus.Errorf("An error occurred while stopping Firecracker VMM: %v", err)
				}
				logrus.Infof("exiting from sigquit")
				os.Exit(0)
			}
		}
	}()
}

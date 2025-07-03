package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/app"
	"strings"
)

func main() {

	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logrus.Fatalf("failed to open log file: %v", err)
	}
	logrus.SetOutput(logFile)

	vmManager := app.NewVMManager()
	app.InstallSignalHandlers(vmManager)

	scanner := bufio.NewScanner(os.Stdin)
	var cmd string

	for {
		fmt.Printf("VM$ ")
		if !scanner.Scan() {
			break // EOF or error
		}
		cmd = strings.ToLower(strings.TrimSpace(scanner.Text()))

		if cmd == "quit" {
			vmManager.GracefulShutdownAll()
			break
		}

		if cmd == "run" {
			id, err := vmManager.CreateVM()
			fmt.Printf("Created VM with ID: %s\n", id.String())

			if err != nil {
				logrus.Errorf("createvm failed: %v", err)
			}
		}
	}

}

package main

import (
	"bufio"
	"fmt"
	"os"

	// "net/http"
	// "os"
	"strings"
	// flags "github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

const (
	// executableMask is the mask needed to check whether or not a file's
	// permissions are executable.
	executableMask = 0111

	firecrackerDefaultPath = "firecracker"
)

func main() {
	
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logrus.Fatalf("failed to open log file: %v", err)
	}
	logrus.SetOutput(logFile)
	
	vmManager := NewVMManager()
	InstallSignalHandlers(vmManager)

	scanner := bufio.NewScanner(os.Stdin)
	var cmd string

	for {
		fmt.Printf("VM$ ")
		if !scanner.Scan() {
			break // EOF or error
		}
		cmd = strings.ToLower(strings.TrimSpace(scanner.Text()))

		if cmd == "quit" {
			break
		}
		if cmd == "run" {
			err := vmManager.CreateVM()
			if err != nil {
				logrus.Errorf("createvm failed")
			}
		}
	}

}

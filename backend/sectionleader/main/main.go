package main

import (
	"bufio"
	"fmt"
	"os"

	// "net/http"
	// "os"
	"strings"
	// flags "github.com/jessevdk/go-flags"
	// log "github.com/sirupsen/logrus"
)

const (
	// executableMask is the mask needed to check whether or not a file's
	// permissions are executable.
	executableMask = 0111

	firecrackerDefaultPath = "firecracker"
)

// var Opts = newOptions();

func main() {
	// r := NewRouter();
	// if err := http.ListenAndServe(":6123", r); err != nil {
	// 	log.Fatalf(err.Error())
	// }

	vmMan := NewVMManager()
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
			// TODO: sync, not sure why using goroutine breaks this
			go vmMan.CreateVM()
		}
	}

}

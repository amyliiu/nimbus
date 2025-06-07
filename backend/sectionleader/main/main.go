// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

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
			go SpawnVM()
		}
	}

}

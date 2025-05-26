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
	"fmt"
	"net/http"
	"os"

	flags "github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

const (
	// executableMask is the mask needed to check whether or not a file's
	// permissions are executable.
	executableMask = 0111

	firecrackerDefaultPath = "firecracker"
)

// var Opts = newOptions();

func main() {
	// opts := newOptions()
	// p := flags.NewParser(opts, flags.Default)
	// // if no args just print help
	// if len(os.Args) == 1 {
	// 	p.WriteHelp(os.Stderr)
	// 	os.Exit(0)
	// }
	// _, err := p.ParseArgs(os.Args)
	// if err != nil {
	// 	// ErrHelp indicates that the help message was printed so we
	// 	// can exit
	// 	if val, ok := err.(*flags.Error); ok && val.Type == flags.ErrHelp {
	// 		os.Exit(0)
	// 	}
	// 	p.WriteHelp(os.Stderr)
	// 	os.Exit(1)
	// }

	// if opts.Version {
	// 	// TODO: placeholder
	// 	fmt.Println("PLACEHOLDER")
	// 	os.Exit(0)
	// }

	// Opts = opts;
	// defer opts.Close()


	// if err := runVMM(context.Background(), opts); err != nil {
	// 	log.Fatalf(err.Error())
	// }
	
	r := NewRouter();
	if err := http.ListenAndServe(":6123", r); err != nil {
		log.Fatalf(err.Error())
	}
}

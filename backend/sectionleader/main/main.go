package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/server"
)

const (
	// executableMask is the mask needed to check whether or not a file's
	// permissions are executable.
	executableMask = 0111

	firecrackerDefaultPath = "firecracker"
)

func main() {
	opts := newOptions()
	defer opts.Close()

	r := server.NewRouter()
	addr := ":8080"
	log.Printf("Starting server at http://localhost%s\n", addr)
	go httpServeLoop(&addr, &r)
	go runVMMLoop(opts)
}

func httpServeLoop(addr *string, r *http.Handler) {
	if err := http.ListenAndServe(*addr, *r); err != nil {
		log.Println(err)
	}
}

func runVMMLoop(opts *options) {
	if err := runVMM(context.Background(), opts); err != nil {
		log.Fatalf(err.Error())
	}
}

// Run a vmm with a given set of options
func runVMM(ctx context.Context, opts *options) error {
	// convert options to a firecracker config
	fcCfg, err := opts.getFirecrackerConfig()
	if err != nil {
		log.Println("Error: %s", err)
		return err
	}

	vmmCtx, vmmCancel := context.WithCancel(ctx)
	defer vmmCancel()

	machineOpts := []firecracker.Opt{
		firecracker.WithProcessRunner(exec.Command("firecracker")),
	}

	var firecrackerBinary string
	firecrackerBinary, err = exec.LookPath(firecrackerDefaultPath)
	if err != nil {
		return err
	}

	finfo, err := os.Stat(firecrackerBinary)
	if os.IsNotExist(err) {
		return fmt.Errorf("Binary %q does not exist: %v", firecrackerBinary, err)
	}

	if err != nil {
		return fmt.Errorf("Failed to stat binary, %q: %v", firecrackerBinary, err)
	}

	if finfo.IsDir() {
		return fmt.Errorf("Binary, %q, is a directory", firecrackerBinary)
	} else if finfo.Mode()&executableMask == 0 {
		return fmt.Errorf("Binary, %q, is not executable. Check permissions of binary", firecrackerBinary)
	}

	// if the jailer is used, the final command will be built in NewMachine()
	if fcCfg.JailerCfg == nil {
		cmd := firecracker.VMCommandBuilder{}.
			WithBin(firecrackerBinary).
			WithSocketPath(fcCfg.SocketPath).
			WithStdin(os.Stdin).
			WithStdout(os.Stdout).
			WithStderr(os.Stderr).
			Build(ctx)

		machineOpts = append(machineOpts, firecracker.WithProcessRunner(cmd))
	}

	m, err := firecracker.NewMachine(vmmCtx, fcCfg, machineOpts...)
	if err != nil {
		return fmt.Errorf("Failed creating machine: %s", err)
	}

	if err := m.Start(vmmCtx); err != nil {
		return fmt.Errorf("Failed to start machine: %v", err)
	}
	defer func() {
		if err := m.StopVMM(); err != nil {
			log.Println("An error occurred while stopping Firecracker VMM: %v", err)
		}
	}()

	if opts.validMetadata != nil {
		if err := m.SetMetadata(vmmCtx, opts.validMetadata); err != nil {
			log.Println("An error occurred while setting Firecracker VM metadata: %v", err)
		}
	}

	installSignalHandlers(vmmCtx, m)

	// wait for the VMM to exit
	if err := m.Wait(vmmCtx); err != nil {
		return fmt.Errorf("Wait returned an error %s", err)
	}
	log.Printf("Start machine was happy")
	return nil
}

// Install custom signal handlers:
func installSignalHandlers(ctx context.Context, m *firecracker.Machine) {
	go func() {
		// Clear some default handlers installed by the firecracker SDK:
		signal.Reset(os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

		for {
			switch s := <-c; {
			case s == syscall.SIGTERM || s == os.Interrupt:
				log.Printf("Caught signal: %s, requesting clean shutdown", s.String())
				if err := m.Shutdown(ctx); err != nil {
					log.Println("An error occurred while shutting down Firecracker VM: %v", err)
				}
			case s == syscall.SIGQUIT:
				log.Printf("Caught signal: %s, forcing shutdown", s.String())
				if err := m.StopVMM(); err != nil {
					log.Println("An error occurred while stopping Firecracker VMM: %v", err)
				}
			}
		}
	}()
}

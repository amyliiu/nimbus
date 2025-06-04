package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"os/exec"
	"os/signal"
	"syscall"

	uuid "github.com/google/uuid"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	log "github.com/sirupsen/logrus"
)

const (
	refSquashFsPath = "./ref/ubuntu-24.04.squashfs.upstream"
	refImgPath      = "./ref/vmlinux"
)

type vmFilePaths struct {
	id            uuid.UUID
	kernelImgPath string
	fsImgPath     string
}

func SpawnVM() error {
	id := uuid.New()
	fmt.Println("New VM created, UUID: ", id.String())

	paths, err := createVMFolder(id)
	if err != nil {
		log.Fatal(err)
		return err
	}

	opts, err := setVMOpts(paths)
	if err != nil {
		log.Fatal(err)
		return err
	}
	// FIXME: idk what this does
	defer opts.Close()

	if err := runVMM(context.Background(), opts); err != nil {
		log.Fatalf("%s", err.Error())
	}
	return nil
}

func createVMFolder(id uuid.UUID) (vmFilePaths, error) {
	err := os.MkdirAll("data/"+id.String(), 0755)
	if err != nil {
		log.Fatal(err)
		return vmFilePaths{}, err
	}
	srcImg, err := os.Open(refImgPath)
	if err != nil {
		return vmFilePaths{}, err
	}
	defer srcImg.Close()
	dstImgPath := "./data/" + id.String() + "/vmlinux"
	dstImg, err := os.Create(dstImgPath)
	if err != nil {
		return vmFilePaths{}, err
	}
	defer dstImg.Close()
	_, err = io.Copy(dstImg, srcImg)
	if err != nil {
		return vmFilePaths{}, err
	}
	srcFs, err := os.Open(refSquashFsPath)
	if err != nil {
		return vmFilePaths{}, err
	}
	defer srcFs.Close()
	dstFsPath := "./data/" + id.String() + "/fs.squashfs.upstream"
	dstFs, err := os.Create(dstFsPath)
	if err != nil {
		return vmFilePaths{}, err
	}
	defer dstFs.Close()

	_, err = io.Copy(dstFs, srcFs)
	if err != nil {
		return vmFilePaths{}, err
	}
	
	extractedFsPath := ""

	return vmFilePaths{id, dstImgPath, extractedFsPath}, nil
}

func setVMOpts(p vmFilePaths) (*options, error) {
	opts := newOptions()
	opts.FcBinary = "firecracker"
	opts.FcKernelImage = p.kernelImgPath
	opts.FcRootDrivePath = p.fsImgPath
	return opts, nil
}

// Run a vmm with a given set of options
func runVMM(ctx context.Context, opts *options) error {
	// convert options to a firecracker config
	fcCfg, err := opts.getFirecrackerConfig()
	if err != nil {
		log.Errorf("Error: %s", err)
		return err
	}
	logger := log.New()

	if opts.Debug {
		log.SetLevel(log.DebugLevel)
		logger.SetLevel(log.DebugLevel)
	}

	vmmCtx, vmmCancel := context.WithCancel(ctx)
	defer vmmCancel()

	machineOpts := []firecracker.Opt{
		firecracker.WithLogger(log.NewEntry(logger)),
	}

	var firecrackerBinary string
	if len(opts.FcBinary) != 0 {
		firecrackerBinary = opts.FcBinary
	} else {
		firecrackerBinary, err = exec.LookPath(firecrackerDefaultPath)
		if err != nil {
			return err
		}
	}

	finfo, err := os.Stat(firecrackerBinary)
	if os.IsNotExist(err) {
		return fmt.Errorf("binary %q does not exist: %v", firecrackerBinary, err)
	}

	if err != nil {
		return fmt.Errorf("failed to stat binary, %q: %v", firecrackerBinary, err)
	}

	if finfo.IsDir() {
		return fmt.Errorf("binary, %q, is a directory", firecrackerBinary)
	} else if finfo.Mode()&executableMask == 0 {
		return fmt.Errorf("binary, %q, is not executable. Check permissions of binary", firecrackerBinary)
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
		return fmt.Errorf("failed creating machine: %s", err)
	}

	if err := m.Start(vmmCtx); err != nil {
		return fmt.Errorf("failed to start machine: %v", err)
	}
	defer func() {
		if err := m.StopVMM(); err != nil {
			log.Errorf("An error occurred while stopping Firecracker VMM: %v", err)
		}
	}()

	if opts.validMetadata != nil {
		if err := m.SetMetadata(vmmCtx, opts.validMetadata); err != nil {
			log.Errorf("An error occurred while setting Firecracker VM metadata: %v", err)
		}
	}

	installSignalHandlers(vmmCtx, m)

	// wait for the VMM to exit
	if err := m.Wait(vmmCtx); err != nil {
		return fmt.Errorf("wait returned an error %s", err)
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
					log.Errorf("An error occurred while shutting down Firecracker VM: %v", err)
				}
			case s == syscall.SIGQUIT:
				log.Printf("Caught signal: %s, forcing shutdown", s.String())
				if err := m.StopVMM(); err != nil {
					log.Errorf("An error occurred while stopping Firecracker VMM: %v", err)
				}
			}
		}
	}()
}

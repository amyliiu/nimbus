package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"os/exec"
	"os/signal"
	"syscall"

	uuid "github.com/google/uuid"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/sirupsen/logrus"
)

const (
	refSquashFsPath = "./ref/squashfs"
	refImgPath      = "./ref/vmlinux"
)

type vmFilePaths struct {
	id            MachineUUID
	kernelImgPath string
	fsRootPath    string
	stdoutPath    string
	stderrPath    string
}

func SpawnNewVM() (*firecracker.Machine, MachineUUID, context.CancelFunc, error) {
	id := MachineUUID(uuid.New())
	fmt.Println("Creating new VM, UUID:", id.String())

	vmPaths, err := createVMFolder(id)
	if err != nil {
		logrus.Fatal(err)
		return nil, id, nil, err
	}

	opts, err := setVMOpts(vmPaths)
	if err != nil {
		logrus.Fatal(err)
		return nil, id, nil, err
	}
	defer opts.Close()

	machine, err := setupFirecrackerMachine(opts)
	if err != nil {
		return nil, id, nil, err
	}
	
	vmmCtx, vmmCancel := context.WithCancel(context.Background())

	go runFirecrackerMachine(vmmCtx, machine)

	return machine, id, vmmCancel, nil
}

func createVMFolder(id MachineUUID) (vmFilePaths, error) {
	dstRootPath := "./data/" + id.String()
	err := os.MkdirAll(dstRootPath, 0755)
	if err != nil {
		logrus.Fatal(err)
		return vmFilePaths{}, err
	}
	srcImg, err := os.Open(refImgPath)
	if err != nil {
		return vmFilePaths{}, err
	}
	defer srcImg.Close()
	dstImgPath := dstRootPath + "/vmlinux"
	dstImg, err := os.Create(dstImgPath)
	if err != nil {
		return vmFilePaths{}, err
	}
	defer dstImg.Close()
	_, err = io.Copy(dstImg, srcImg)
	if err != nil {
		return vmFilePaths{}, err
	}

	extractedFsPath := dstRootPath + "/squashfs-root"
	err = exec.Command("unsquashfs", "-d", extractedFsPath, refSquashFsPath).Run()
	if err != nil {
		logrus.Fatal(err)
		return vmFilePaths{}, err
	}

	stdoutPath := dstRootPath + "/log/stdout.log"
	os.MkdirAll(filepath.Dir(stdoutPath), 0555)
	os.Create(stdoutPath)

	stderrPath := dstRootPath + "/log/stderr.log"
	os.MkdirAll(filepath.Dir(stderrPath), 0555)
	os.Create(stderrPath)

	err = exec.Command("./prepVM.sh", dstRootPath).Run()
	if err != nil {
		logrus.Fatal(err)
		return vmFilePaths{}, err
	}

	fsExt4Path := dstRootPath + "/fs.ext4"

	return vmFilePaths{id, dstImgPath, fsExt4Path, stdoutPath, stderrPath}, nil
}

func setVMOpts(p vmFilePaths) (*options, error) {
	opts := newOptions()
	opts.FcBinary = "../../firecracker/release/firecracker"
	opts.FcKernelImage = p.kernelImgPath
	opts.FcRootDrivePath = p.fsRootPath
	opts.FcCPUCount = 1
	opts.FcMemSz = 512
	opts.FcSocketPath = "/tmp/firecracker-" + p.id.String() + ".socket"
	CniNetworkName, err := GenerateCniConfFile(p.id)
	if err != nil {
		return nil, err
	}
	opts.CniNetworkName = CniNetworkName
	// opts.FcNicConfig = []string{"tap0/06:00:AC:10:00:02"}
	opts.FcStdoutPath = p.stdoutPath
	opts.FcStderrPath = p.stderrPath
	return opts, nil
}

func runFirecrackerMachine(ctx context.Context, m *firecracker.Machine) {
	if err := m.Start(ctx); err != nil {
		logrus.Errorf("failed to start machine: %v", err)
		return
	}
	defer func() {
		if err := m.StopVMM(); err != nil {
			logrus.Errorf("An error occurred while stopping Firecracker VMM: %v", err)
		}
	}()

	installSignalHandlers(ctx, m)

	// wait for the VMM to exit
	if err := m.Wait(ctx); err != nil {
		logrus.Errorf("wait returned error %v", err)
	}
	logrus.Printf("Start machine was happy")
}

// Run a vmm with a given set of options
func setupFirecrackerMachine(opts *options) (*firecracker.Machine, error) {
	// convert options to a firecracker config
	fcCfg, err := opts.getFirecrackerConfig()
	if err != nil {
		logrus.Errorf("Error: %s", err)
		return nil, err
	}

	machineOpts := []firecracker.Opt{
		firecracker.WithLogger(logrus.NewEntry(logrus.StandardLogger())),
	}

	var firecrackerBinary string
	if len(opts.FcBinary) != 0 {
		firecrackerBinary = opts.FcBinary
	} else {
		firecrackerBinary, err = exec.LookPath(firecrackerDefaultPath)
		if err != nil {
			return nil, err
		}
	}

	finfo, err := os.Stat(firecrackerBinary)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("binary %q does not exist: %v", firecrackerBinary, err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to stat binary, %q: %v", firecrackerBinary, err)
	}

	if finfo.IsDir() {
		return nil, fmt.Errorf("binary, %q, is a directory", firecrackerBinary)
	} else if finfo.Mode()&executableMask == 0 {
		return nil, fmt.Errorf("binary, %q, is not executable. Check permissions of binary", firecrackerBinary)
	}

	// if the jailer is used, the final command will be built in NewMachine()
	if fcCfg.JailerCfg == nil {
		stdoutFile, err := os.OpenFile(opts.FcStdoutPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open stdout file %s: %v", opts.FcStdoutPath, err)
		}
		
		stderrFile, err := os.OpenFile(opts.FcStderrPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open stderr file %s: %v", opts.FcStderrPath, err)
		}

		cmd := firecracker.VMCommandBuilder{}.
			WithBin(firecrackerBinary).
			WithSocketPath(fcCfg.SocketPath).
			// WithStdin(os.Stdin).
			WithStdout(stdoutFile).
			WithStderr(stderrFile).
			Build(context.Background())

		machineOpts = append(machineOpts, firecracker.WithProcessRunner(cmd))
	}

	m, err := firecracker.NewMachine(context.Background(), fcCfg, machineOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed creating machine: %s", err)
	}
	
	return m, nil
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
				logrus.Printf("Caught signal: %s, requesting clean shutdown", s.String())
				if err := m.Shutdown(ctx); err != nil {
					logrus.Errorf("An error occurred while shutting down Firecracker VM: %v", err)
				}
			case s == syscall.SIGQUIT:
				logrus.Printf("Caught signal: %s, forcing shutdown", s.String())
				if err := m.StopVMM(); err != nil {
					logrus.Errorf("An error occurred while stopping Firecracker VMM: %v", err)
				}
			}
		}
	}()
}

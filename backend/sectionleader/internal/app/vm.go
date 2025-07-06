package app

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/sirupsen/logrus"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/constants"
)

type VMState int

const (
	StateActive VMState = iota
	StatePaused
	StateStopped
)

type MachineData struct {
	Id           MachineUUID
	Name         string
	LocalIp      net.IPNet
	RemotePort   int
	CreationTime time.Time
}

type VM struct {
	Machine *firecracker.Machine
	Id      MachineUUID
	State   VMState
	cancel  context.CancelFunc
	data    MachineData
}

type VMManager struct {
	mutex         sync.Mutex
	createVmMutex sync.Mutex
	idNameMap     *IdNameMap
	VMs           map[MachineUUID]*VM
}

func NewVMManager() *VMManager {
	return &VMManager{
		mutex:         sync.Mutex{},
		createVmMutex: sync.Mutex{},
		idNameMap:     NewIdNameMap(),
		VMs:           make(map[MachineUUID]*VM),
	}
}

func (manager *VMManager) CreateVM() (<-chan *MachineData, error) {
	// has to be withcancel as this is the context that lives with the machine
	ctx, cancelFunc := context.WithCancel(context.Background())
	outputChannel := make(chan *MachineData)

	go func() {
		manager.createVmMutex.Lock()
		defer manager.createVmMutex.Unlock()

		machine, id, ip, err := SpawnNewVM(ctx)
		if err != nil {
			logrus.Errorf("failed to spawn VM: %v", err)
			outputChannel <- nil
			return
		}

		if machine == nil {
			logrus.Errorf("spawnvm return error")
			outputChannel <- nil
			return
		}

		manager.mutex.Lock()
		defer manager.mutex.Unlock()

		vmName, err := manager.idNameMap.GenerateNewName(id)
		if err != nil {
			logrus.Errorf("could not generate name for new vm: %v", err)
			return
		}

		vmPtr := &VM{
			Machine: machine,
			Id:      id,
			State:   StateActive,
			cancel:  cancelFunc,
			data: MachineData{
				Id:           id,
				Name:         vmName,
				LocalIp:      ip,
				CreationTime: time.Now(),
				RemotePort:   constants.MinRemotePort + len(manager.VMs),
			}}

		manager.VMs[id] = vmPtr
		outputChannel <- &vmPtr.data
	}()

	return outputChannel, nil
}

func (manager *VMManager) PauseVM(id MachineUUID) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)

	go func(ctx context.Context, cancelFunc context.CancelFunc, manager *VMManager, id MachineUUID) {
		defer cancelFunc()

		manager.mutex.Lock()
		defer manager.mutex.Unlock()

		vmPtr := manager.VMs[id]
		if vmPtr.State != StateActive {
			logrus.Errorf("machine not active, cannot be paused, id: %s", id.String())
			return
		}

		err := vmPtr.Machine.PauseVM(ctx)
		if err != nil {
			logrus.Errorf("pause vm error")
			return
		}

		vmPtr.State = StatePaused
	}(ctx, cancelFunc, manager, id)
}

func (manager *VMManager) ResumeVM(id MachineUUID) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)

	go func(ctx context.Context, cancelFunc context.CancelFunc, manager *VMManager, id MachineUUID) {
		defer cancelFunc()

		vmPtr := manager.VMs[id]
		if vmPtr.State != StatePaused {
			logrus.Errorf("machine not paused, cannot be resumed, id %s", id.String())
			return
		}
		err := vmPtr.Machine.PauseVM(ctx)
		if err != nil {
			logrus.Errorf("resume vm error")
			return
		}

		vmPtr.State = StateActive
	}(ctx, cancelFunc, manager, id)
}

func (manager *VMManager) GracefulShutdownVM(id MachineUUID) <-chan bool {
	ctx, cancelFunc := context.WithTimeout(context.Background(), constants.DefaultTimeout)
	logrus.Infof("requested machine %s shutdown", id.String())
	outputChan := make(chan bool)

	go func() {
		defer cancelFunc()
		defer close(outputChan)

		manager.mutex.Lock()
		defer manager.mutex.Unlock()

		vmPtr := manager.VMs[id]
		defer vmPtr.cancel()

		if vmPtr.State == StateStopped {
			logrus.Errorf("attempted to shutdown stopped machine, id: %s", id.String())
			return
		}

		vmPtr.State = StateStopped
		err := vmPtr.Machine.Shutdown(ctx)
		if err != nil {
			logrus.Errorf("machine shutdown err, id: %s, err %v, forcing shutdown", id.String(), err)
			if forceErr := vmPtr.Machine.StopVMM(); forceErr != nil {
				logrus.Errorf("force shutdown failed, id: %s, err %v", id.String(), forceErr)
				outputChan <- false
				return
			}
			return
		}

		outputChan <- true
		logrus.Infof("machine %s successfully shut down", id.String())
	}()

	return outputChan
}

func (manager *VMManager) GracefulShutdownAll() error {
	shutdownChans := make([]<-chan bool, len(manager.VMs))

	counter := 0
	for id := range manager.VMs {
		shutdownChans[counter] = manager.GracefulShutdownVM(id)
		counter++
	}

	for _, c := range shutdownChans {
		select {
		case <-time.After(constants.DefaultTimeout):
			logrus.Errorf("vm shutdown timeout")
			return fmt.Errorf("vm shutdown timeout")
		case <-c:
			// correct, proceed
		}
	}

	return nil
}

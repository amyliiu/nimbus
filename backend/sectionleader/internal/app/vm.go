package app

import (
	"context"
	"net"
	"sync"
	"time"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/sirupsen/logrus"
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
	VMs           map[MachineUUID]*VM
}

func NewVMManager() *VMManager {
	return &VMManager{
		mutex:         sync.Mutex{},
		createVmMutex: sync.Mutex{},
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

		vmPtr := &VM{
			Machine: machine,
			Id:      id,
			State:   StateActive,
			cancel:  cancelFunc,
			data: MachineData{
				Id:           id,
				Name:         "placeholder",
				LocalIp:      ip,
				CreationTime: time.Now(),
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

func (manager *VMManager) GracefulShutdownVM(id MachineUUID) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)

	go func(ctx context.Context, cancelFunc context.CancelFunc, manager *VMManager, id MachineUUID) {
		defer cancelFunc()

		manager.mutex.Lock()
		defer manager.mutex.Unlock()

		vmPtr := manager.VMs[id]
		defer vmPtr.cancel()

		if vmPtr.State == StateStopped {
			logrus.Errorf("attempted to shutdown stopped machine, id: %s", id.String())
			return
		}

		vmPtr.State = StateStopped
		err := vmPtr.Machine.Shutdown(context.Background())
		if err != nil {
			logrus.Errorf("machine shutdown err, id: %s, err %v, forcing shutdown", id.String(), err)
			return
		}
	}(ctx, cancelFunc, manager, id)
}

func (manager *VMManager) GracefulShutdownAll() error {
	for id := range manager.VMs {
		manager.GracefulShutdownVM(id)
	}
	return nil
}

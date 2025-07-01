package main

import (
	"context"
	"fmt"
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

type VM struct {
	Machine *firecracker.Machine
	Id      MachineUUID
	State   VMState
	cancel  context.CancelFunc
}

type VMManager struct {
	mutex sync.RWMutex
	VMs   map[MachineUUID]*VM
}

func NewVMManager() *VMManager {
	return &VMManager{
		mutex: sync.RWMutex{},
		VMs:   make(map[MachineUUID]*VM),
	}
}

func (manager *VMManager) CreateVM() error {
	machine, id, cancelFunc, err := SpawnNewVM()
	if err != nil {
		logrus.Errorf("failed to spawn VM: %v", err)
		return err
	}
	if cancelFunc == nil || machine == nil {
		return fmt.Errorf("spawnvm return error")
	}
	vmPtr := &VM{
		Machine: machine,
		Id:      id,
		State:   StateActive,
		cancel:  cancelFunc,
	}

	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	manager.VMs[id] = vmPtr
	return nil
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

func (manager *VMManager) ResumeVM(id MachineUUID) error {
	vmPtr := manager.VMs[id]
	if vmPtr.State != StatePaused {
		return fmt.Errorf("machine not paused, cannot be resumed, id %s", id.String())
	}
	vmPtr.State = StateActive

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)

	go func(ctx context.Context, cancelFunc context.CancelFunc, vmPtr *VM) {
		defer cancelFunc()
		err := vmPtr.Machine.PauseVM(ctx)
		if err != nil {
			logrus.Errorf("pause vm error")
		}
	}(ctx, cancelFunc, vmPtr)

	return nil
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

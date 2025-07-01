package main

import (
	"context"
	"fmt"
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
	State  VMState
	cancel  context.CancelFunc
}

type VMManager struct {
	VMs map[MachineUUID]*VM
}

func NewVMManager() *VMManager {
	return &VMManager{
		VMs: make(map[MachineUUID]*VM),
	}
}

func (manager *VMManager) CreateVM() error {
	machine, id, cancelFunc, err := SpawnNewVM()
	if err != nil {
		logrus.Errorf("failed to spawn VM: %v", err)
		return err
	}
	if cancelFunc == nil || machine == nil{
		return fmt.Errorf("spawnvm return error")
	}
	vmPtr := &VM{
		Machine: machine,
		Id:      id,
		State:  StateActive,
		cancel:  cancelFunc,
	}
	manager.VMs[id] = vmPtr
	return nil
}

func (manager *VMManager) PauseVM(id MachineUUID) error {
	vmPtr := manager.VMs[id]
	if vmPtr.State != StateActive {
		return fmt.Errorf("machine not active, cannot be paused")	
	}
	vmPtr.State = StatePaused
	
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second * 5)

	go func(ctx context.Context, cancelFunc context.CancelFunc, vmPtr *VM) {
		defer cancelFunc()
		err := vmPtr.Machine.PauseVM(ctx)
		if err != nil {
			logrus.Errorf("pause vm error")
		}
	}(ctx, cancelFunc,vmPtr)
	
	return nil
}

func (manager *VMManager) GracefulShutdownVM(id MachineUUID) error {
	vmPtr := manager.VMs[id]
	err := vmPtr.Machine.Shutdown(context.Background())
	if err != nil {
		return err
	}

	vmPtr.cancel()
	vmPtr.State = StateStopped

	return nil
}

func (manager *VMManager) GracefulShutdownAll() error {
	for id := range manager.VMs {
		err := manager.GracefulShutdownVM(id)
		if err != nil {
			return err
		}
	}
	return nil
}
package main

import (
	"context"
	"fmt"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/sirupsen/logrus"
)

type VM struct {
	Machine *firecracker.Machine
	Id      MachineUUID
	Active  bool
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

func (man *VMManager) CreateVM() error {
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
		Active:  true,
		cancel:  cancelFunc,
	}
	man.VMs[id] = vmPtr
	return nil
}

func (man *VMManager) RemoveVM() {

}

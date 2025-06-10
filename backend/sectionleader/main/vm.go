package main

import (
	"context"

	"github.com/firecracker-microvm/firecracker-go-sdk"
)

type VM struct {
	Machine *firecracker.Machine
	Id      MachineUUID
	Active  bool
	ctx     context.Context
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
	vmCtx, vmCancel := context.WithCancel(context.Background())
	machine, id, err := SpawnVM(vmCtx)
	if err != nil {
		return err
	}
	vm := &VM{
		Machine: machine,
		Id:      id,
		Active:  true,
		ctx:     vmCtx,
		cancel:  vmCancel,
	}
	man.VMs[id] = vm
	return nil
}

func (man *VMManager) RemoveVM() {

}

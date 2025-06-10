package main

import (
	"github.com/firecracker-microvm/firecracker-go-sdk"
)


type VM struct {
	Machine *firecracker.Machine
	Id MachineUUID
	Active bool
}

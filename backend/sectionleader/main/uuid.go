package main

import "github.com/google/uuid"

type MachineUUID uuid.UUID

func (o MachineUUID) String() string {
	return uuid.UUID(o).String()
}

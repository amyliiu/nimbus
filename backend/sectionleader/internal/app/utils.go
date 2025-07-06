package app

import (
	"fmt"

	"github.com/google/uuid"
)

type MachineUUID uuid.UUID

func (o MachineUUID) String() string {
	return uuid.UUID(o).String()
}

type IdNameMap struct {
	idToName map[MachineUUID]string
	nameToId map[string]MachineUUID
}

func NewIdNameMap() *IdNameMap {
	return &IdNameMap{
		idToName: make(map[MachineUUID]string),
		nameToId: make(map[string]MachineUUID),
	}
}

func (m *IdNameMap) GenerateNewName(id MachineUUID) (string,error) {
	name := GeneratePetname()
	_, ok := m.nameToId[name]

	counter := 0 
	for ok && counter < 10 {
		name = GeneratePetname()
		_, ok = m.nameToId[name]
		counter++
	}
	if counter == 9 {
		return "", fmt.Errorf("could not generate new name")
	}
	
	m.idToName[id] = name
	m.nameToId[name] = id
	return name, nil
}

func (m *IdNameMap) GetName(id MachineUUID) (string, error) {
	name, ok := m.idToName[id]
	if ok {
		return name, nil
	}

	return "", fmt.Errorf("id not found")
}

func (m *IdNameMap) GetId(name string) (MachineUUID, error) {
	id, ok := m.nameToId[name]
	if ok {
		return id, nil
	}

	return MachineUUID{}, fmt.Errorf("id not found")
	
}


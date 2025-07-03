package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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


func InstallSignalHandlers(manager *VMManager) {
	go func() {
		// Clear some default handlers installed by the firecracker SDK:
		signal.Reset(os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

		for {
			switch s := <-c; {
			case s == syscall.SIGTERM || s == os.Interrupt:
				logrus.Printf("Caught signal: %s, requesting clean shutdown", s.String())
				err := manager.GracefulShutdownAll()
				if err != nil {
					logrus.Errorf("An error occurred while stopping Firecracker VMM: %v", err)
				}
			case s == syscall.SIGQUIT:
				// FIXME: force shutdown
				logrus.Printf("Caught signal: %s, forcing shutdown", s.String())
				// if err := m.StopVMM(); err != nil {
				// 	logrus.Errorf("An error occurred while stopping Firecracker VMM: %v", err)
				// }
				err := manager.GracefulShutdownAll()
				if err != nil {
					logrus.Errorf("An error occurred while stopping Firecracker VMM: %v", err)
				}
			}
			println()
			os.Exit(0)
		}
	}()
}
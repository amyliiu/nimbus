package main

import (
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
package app

import (
	"fmt"
	"net"
	"os/exec"
	"strconv"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/constants"
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

// SetupPortForwarding creates iptables rules to forward traffic from internal VM port to local host port
func SetupPortForwarding(vmIP net.IP, localPort int) error {
	if localPort < constants.MinLocalForwardPort || localPort > constants.MaxLocalForwardPort {
		return fmt.Errorf("local port %d outside allowed range (%d-%d)", 
			localPort, constants.MinLocalForwardPort, constants.MaxLocalForwardPort)
	}

	// Forward traffic from localhost:localPort to vmIP:25565
	// DNAT rule: redirect incoming traffic on localPort to VM's port 25565
	dnatRule := []string{
		"-t", "nat",
		"-A", "OUTPUT", 
		"-p", "tcp",
		"--dport", strconv.Itoa(localPort),
		"-d", "127.0.0.1",
		"-j", "DNAT",
		"--to-destination", fmt.Sprintf("%s:%d", vmIP.String(), constants.InternalGamePort),
	}

	// Forward traffic from external interfaces to VM
	prerouting := []string{
		"-t", "nat",
		"-A", "PREROUTING",
		"-p", "tcp",
		"--dport", strconv.Itoa(localPort),
		"-j", "DNAT", 
		"--to-destination", fmt.Sprintf("%s:%d", vmIP.String(), constants.InternalGamePort),
	}

	// Allow forwarding in FORWARD chain
	forwardRule := []string{
		"-A", "FORWARD",
		"-p", "tcp",
		"-d", vmIP.String(),
		"--dport", strconv.Itoa(constants.InternalGamePort),
		"-j", "ACCEPT",
	}

	// SNAT for return traffic
	snatRule := []string{
		"-t", "nat",
		"-A", "POSTROUTING",
		"-p", "tcp",
		"-s", vmIP.String(),
		"--sport", strconv.Itoa(constants.InternalGamePort),
		"-j", "MASQUERADE",
	}

	// Execute iptables rules
	rules := [][]string{dnatRule, prerouting, forwardRule, snatRule}
	
	for _, rule := range rules {
		cmd := exec.Command("iptables", rule...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			logrus.Errorf("iptables rule failed: %v, output: %s, rule: %v", err, output, rule)
			return fmt.Errorf("failed to add iptables rule: %v", err)
		}
		logrus.Infof("Added iptables rule: %v", rule)
	}

	logrus.Infof("Successfully set up port forwarding from localhost:%d to %s:%d", 
		localPort, vmIP.String(), constants.InternalGamePort)
	return nil
}

// CleanupPortForwarding removes iptables rules for a specific VM
func CleanupPortForwarding(vmIP net.IP, localPort int) error {
	// Remove the rules by changing -A to -D
	dnatRule := []string{
		"-t", "nat",
		"-D", "OUTPUT", 
		"-p", "tcp",
		"--dport", strconv.Itoa(localPort),
		"-d", "127.0.0.1",
		"-j", "DNAT",
		"--to-destination", fmt.Sprintf("%s:%d", vmIP.String(), constants.InternalGamePort),
	}

	prerouting := []string{
		"-t", "nat",
		"-D", "PREROUTING",
		"-p", "tcp",
		"--dport", strconv.Itoa(localPort),
		"-j", "DNAT", 
		"--to-destination", fmt.Sprintf("%s:%d", vmIP.String(), constants.InternalGamePort),
	}

	forwardRule := []string{
		"-D", "FORWARD",
		"-p", "tcp",
		"-d", vmIP.String(),
		"--dport", strconv.Itoa(constants.InternalGamePort),
		"-j", "ACCEPT",
	}

	snatRule := []string{
		"-t", "nat",
		"-D", "POSTROUTING",
		"-p", "tcp",
		"-s", vmIP.String(),
		"--sport", strconv.Itoa(constants.InternalGamePort),
		"-j", "MASQUERADE",
	}

	// Execute cleanup rules
	rules := [][]string{dnatRule, prerouting, forwardRule, snatRule}
	
	for _, rule := range rules {
		cmd := exec.Command("iptables", rule...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			logrus.Warnf("iptables cleanup rule failed (may not exist): %v, output: %s, rule: %v", err, output, rule)
			// Don't return error for cleanup failures - rules might not exist
		} else {
			logrus.Infof("Removed iptables rule: %v", rule)
		}
	}

	logrus.Infof("Cleaned up port forwarding for %s:%d -> localhost:%d", 
		vmIP.String(), constants.InternalGamePort, localPort)
	return nil
}


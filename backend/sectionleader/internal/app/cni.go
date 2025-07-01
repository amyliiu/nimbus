package app

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

const CniConfRootDir = "/etc/cni/conf.d"

type IPAM struct {
	Type   string  `json:"type"`
	Subnet string  `json:"subnet"`
	Routes []Route `json:"routes"`
}

type Route struct {
	Dst string `json:"dst"`
}

type Plugin struct {
	Type string `json:"type"`
	IPAM *IPAM  `json:"ipam,omitempty"`
}

type CNIConfig struct {
	CNIVersion string   `json:"cniVersion"`
	Name       string   `json:"name"`
	Plugins    []Plugin `json:"plugins"`
}

var usedSubnets map[uint8]struct{} = make(map[uint8]struct{})

// returns name of the config generated
func GenerateCniConfFile(id MachineUUID) (string, error) {
	vmID := id.String()
	// generate unused subnet id
	subnetID := uint8(0)
	for i := uint8(1); i <= 254; i++ {
		if _, used := usedSubnets[i]; !used {
			subnetID = i
			usedSubnets[i] = struct{}{}
			break
		}
	}
	if subnetID == 0 {
		return "", fmt.Errorf("ran out of subnet IDs")
	}

	subnet := fmt.Sprintf("192.168.%d.0/24", subnetID)
	config := CNIConfig{
		CNIVersion: "0.4.0",
		Name:       fmt.Sprintf("fcnet-%s", vmID),
		Plugins: []Plugin{
			{
				Type: "ptp",
				IPAM: &IPAM{
					Type:   "host-local",
					Subnet: subnet,
					Routes: []Route{
						{Dst: "0.0.0.0/0"},
					},
				},
			},
			{
				Type: "firewall",
			},
			{
				Type: "tc-redirect-tap",
			},
		},
	}

	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}

	confPath := CniConfRootDir + "/fcnet-" + id.String() + ".conflist"
	err = os.MkdirAll(CniConfRootDir, 0755)
	if err != nil {
		return "", err
	}

	logrus.Infof("created CNI config, with subnet %s, path %s", subnet, confPath)
	return config.Name, os.WriteFile(confPath, jsonBytes, 0644)
}

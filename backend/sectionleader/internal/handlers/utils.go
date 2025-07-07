package handlers

import (
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/BurntSushi/toml"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/app"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/constants"
)

type proxyConfig struct {
	Name       string `toml:"name"`
	ConnType   string `toml:"type"`
	LocalIp    net.IP `toml:"localIP"`
	LocalPort  int    `toml:"localPort"`
	RemotePort int    `toml:"remotePort"`
}

type frpcConfig struct {
	Proxies []proxyConfig `toml:"proxies"`
}

func CreateTomlFrpcConfig(data *app.MachineData) error {
	if data.RemotePort < constants.MinRemotePort || data.RemotePort > constants.MaxRemotePort {
		return fmt.Errorf("port requested outside allowed port range")
	}
	cfg := proxyConfig{
		Name:       data.Id.String(),
		ConnType:   "tcp",
		LocalIp:    data.LocalIp.IP,
		LocalPort:  22,
		RemotePort: data.RemotePort,
	}

	proxiesConfig := frpcConfig{
		Proxies: []proxyConfig{cfg},
	}

	err := os.MkdirAll(constants.FrpcConfigDir, 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(constants.FrpcConfigDir + "/" + data.Id.String() + ".toml")
	if err != nil {
		return err
	}
	defer file.Close()

	err = toml.NewEncoder(file).Encode(proxiesConfig)
	if err != nil {
		return err
	}

	cmd := exec.Command("su", "tswu", "-c", constants.RefreshFrpcPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("reload frpc error, err: %v output: %s", err, output)
	}

	return nil
}

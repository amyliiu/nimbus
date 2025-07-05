package handlers

import (
	"bytes"
	"fmt"
	"net"

	"github.com/BurntSushi/toml"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/app"
	"github.com/tongshengw/nimbus/backend/sectionleader/internal/constants"
)

type proxyConfig struct {
	Name       string
	ConnType   string
	LocalIp    net.IP
	LocalPort  int
	RemotePort int
}

type frpcConfig struct {
	Proxies []proxyConfig
}

func CreateTomlFrpcConfig(data *app.MachineData, remotePort int) error {
	if remotePort < constants.MinRemotePort || remotePort > constants.MaxRemotePort {
		return fmt.Errorf("port requested outside allowed port range")
	}
	cfg := proxyConfig{
		Name:       data.Id.String(),
		ConnType:   "tcp",
		LocalIp:    data.LocalIp.IP,
		LocalPort:  22,
		RemotePort: remotePort,
	}

	proxiesConfig := frpcConfig{
		Proxies: []proxyConfig{cfg},
	}
	
	strRes := bytes.Buffer{}

	err := toml.NewEncoder(&strRes).Encode(proxiesConfig)
	if err != nil {
		return err
	}
	
	// FIXME: change to write to file
	fmt.Printf("%v", strRes.String())
	
	return nil
}

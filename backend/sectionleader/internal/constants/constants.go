package constants

import (
	"time"
)

const (
	DefaultTimeout  = time.Second * 5
	CreateVmTimeout = time.Second * 20

	FrpcPath      = "/home/linuxbrew/.linuxbrew/bin/frpc"
	FrpcConfigDir = "/home/tswu/frpc/nimbus"

	RefreshFrpcPath = "/home/tswu/refresh-frpc.sh"

	MinRemotePort = 8000
	MaxRemotePort = 9000

	// Port forwarding constants for VM internal port 25565
	MinLocalForwardPort = 10000
	MaxLocalForwardPort = 11000
	MinGameRemotePort   = 12000
	MaxGameRemotePort   = 13000
	InternalGamePort    = 25565

	DataDirPath = "./_data"

	CniFirstSubnetStr = "192.168.1.0"
	CniLastSubnetStr = "192.168.254.0"
	PublicIpStr = "18.119.116.39"
)

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

	DataDirPath = "./_data"

	CniFirstSubnetStr = "192.168.1.0"
	CniLastSubnetStr = "192.168.254.0"
	PublicIpStr = "18.119.116.39"
)

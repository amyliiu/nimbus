package constants

import (
	"net"
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
)

var PublicIp net.IP = net.ParseIP("18.119.116.39")
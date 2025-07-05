package constants

import "time"

const (
	DefaultTimeout = time.Second * 5
	CreateVmTimeout = time.Second * 20
	
	FrpcConfigDir = "/home/tswu/frpc/nimbus"
	
	MinRemotePort = 8000
	MaxRemotePort = 9000
)
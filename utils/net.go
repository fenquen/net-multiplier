package utils

import "net-multiplier/config"

var localTcpClientPort int32 = 0

func GetLocalTcpClientPort() int32 {
	return GetCyclic(&localTcpClientPort, 1, int32(*config.LocalPortCeil), int32(*config.LocalPortFloor))
}

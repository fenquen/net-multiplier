package utils
var localTcpClientPort int32 = 0

func GetLocalTcpClientPort() int32 {
	return GetCyclic(&localTcpClientPort, 1, 65531, 60000)
}
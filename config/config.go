package config

import "flag"

const TCP_TYPE = "tcp"
const DELIMITER = ","

var (
	LocalTcpSvrAddr = *(flag.String("localTcpSvrAddr",
		"0.0.0.0:9070",
		"the address where the server listens"))

	DestTcpSvrAddrs = *(flag.String("destTcpSvrAddrs",
		"192.168.100.100:8889,192.168.100.100:8888",
		"the destinations that the data is relayed to,it is a comma-delimited string,e.g. 192.168.1.6:9060,192.168.1.60:9060"))

	LocalTcpClientHost = *(flag.String("localTcpClientHost",
		"0.0.0.0",
		"designate the host to which the sender is bind to"))
)

func init() {
	flag.Parse()
}

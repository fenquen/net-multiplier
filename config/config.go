package config

import (
	"flag"
	"strings"
)

const TCP_MODE = "tcp"
const UDP_MODE = "udp"
const DELIMITER = ","

var (
	LocalSvrAddr = *(flag.String("localTcpSvrAddr",
		"0.0.0.0:9070",
		"the address where the server listens"))

	// 192.168.100.100:8889,192.168.100.100:8888
	DestSvrAddrs = *(flag.String("destTcpSvrAddrs",
		"192.168.100.100:8080",
		"the destinations that the data is relayed to,it is a comma-delimited string,e.g. 192.168.1.6:9060,192.168.1.60:9060"))

	LocalClientHost = *(flag.String("localTcpClientHost",
		"0.0.0.0",
		"designate the host to which the sender is bind to"))

	Mode = strings.ToLower(*(flag.String("mode", "udp", "tcp or udp")))

	TempByteSliceLen = *(flag.Int("tempByteSliceLen",2048,"the temp byte slice size for tcp/udp read"))
)

var APP_NAME = "net-multiplier"

func init() {
	flag.Parse()
}

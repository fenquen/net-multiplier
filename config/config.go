package config

import (
	"flag"
)

const TCP_MODE = "tcp"
const UDP_MODE = "udp"
const DELIMITER = ","

var (
	LocalSvrAddr = flag.String("local.svr.addr",
		"192.168.99.60:9070",
		"the address where the server listens")

	// 192.168.100.100:8889,192.168.100.100:8888
	DestSvrAddrs = flag.String("dest.svr.addrs",
		"192.168.99.60:50000",
		"the destinations that the data is relayed to,it is a comma-delimited string,e.g. 192.168.1.6:9060,192.168.1.60:9060")

	LocalClientHost = flag.String("local.client.host",
		"192.168.99.60",
		"designate the host to which the sender is bind to")

	LogLevel = flag.String("log.level", "info", "")

	Mode = flag.String("mode", "udp", "tcp or udp")

	TempByteSliceLen = flag.Int("tempByteSliceLen", 2048,
		"the temp byte slice size for tcp/udp read")

	LocalHttpSvrAddr = flag.String("local.http.svr.addr",
		"192.168.99.60:10060",
		"the http server address where the handle requests")

	LocalPortCeil = flag.Int("local.port.ceil",62000,"")

	LocalPortFloor = flag.Int("local.port.floor",60000,"")
)

var APP_NAME = "net-multiplier"

func init() { flag.Parse() }

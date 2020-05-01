package config

import (
	"flag"
)

const TCP_MODE = "tcp"
const UDP_MODE = "udp"
const DELIMITER = ","

var (
	LocalSvrHostStr = flag.String("local.svr.host",
		"192.168.99.60",
		"the host to which the listener is bind to")

	/*LocalSvrAddr = flag.String("local.svr.addr",
		"192.168.99.60:9070",
		"the address where the server listens")*/

	// 192.168.100.100:8889,192.168.100.100:8888
	/*DestSvrAddrs = flag.String("dest.svr.addrs",
		"192.168.99.60:50000",
		"the destinations that the data is relayed to,it is a comma-delimited string,e.g. 192.168.1.6:9060,192.168.1.60:9060")*/

	LocalClientHostStr = flag.String("local.client.host",
		"192.168.99.60",
		"the host to which the sender is bind to")

	LogLevel = flag.String("log.level", "info", "")

	DefaultMode = flag.String("default_mode", "udp",
		"tcp or udp,it is used only when you add a task without designating an explicit mode")

	DefaultTempByteSliceLen = flag.Int("defaultTempByteSliceLen", 2048,
		"the temp byte slice size for tcp/udp read")

	LocalHttpSvrAddr = flag.String("local.http.svr.addr",
		"192.168.99.60:10060",
		"the http server address where the requests are handled")

	LocalPortCeil = flag.Int("local.port.ceil",62000,"")

	LocalPortFloor = flag.Int("local.port.floor",60000,"")
)

var APP_NAME = "net-multiplier"

func init() { flag.Parse() }

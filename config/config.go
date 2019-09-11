package config

import "flag"

const TCP_TYPE = "tcp"

var (
	LocalTcpSvrAddr = *(flag.String("localTcpSvrAddr", "0.0.0.0:9070", "the address where the server listens"))
	DestTcpSvrAddr  = *(flag.String("destTcpSvrAddr", "", "the destinations that the data is relayed to,it is a comma-delimited string"))
)

func init() {
	flag.Parse()
}

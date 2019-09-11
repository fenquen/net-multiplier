package client

import (
	"net"
	"tcp-multiplier/config"
)

func Send(dataChan chan []byte) {
	localTcpClientAddr, _ := net.ResolveTCPAddr(config.TCP_TYPE, ":9071")
	destTcpSvrAddr, _ := net.ResolveTCPAddr(config.TCP_TYPE, "127.0.0.1:9071")
	tcpConn2Dest, _ := net.DialTCP(config.TCP_TYPE, localTcpClientAddr, destTcpSvrAddr)

	defer func() {
		_ = tcpConn2Dest.Close()
	}()

	for {
		select {
		case byteSlice := <-dataChan:
			_, _ = tcpConn2Dest.Write(byteSlice)
		}
	}

}

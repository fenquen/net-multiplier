package server

import (
	"net"
	"tcp-multiplier/client"
	"tcp-multiplier/config"
)

func ListenAndServe() {
	tempByteSlice := make([]byte, 1024, 1024)
	dataChan := make(chan []byte, 100)
	go client.Send(dataChan)

	for {
		localTcpSvrAddr, err := net.ResolveTCPAddr(config.TCP_TYPE, config.LocalTcpSvrAddr)
		if nil != err {
			panic(err)
		}

		tcpListener, err := net.ListenTCP(config.TCP_TYPE, localTcpSvrAddr)

		srcTcpConn, err := tcpListener.AcceptTCP()


		_, _ = srcTcpConn.Read(tempByteSlice)

		dataChan <- tempByteSlice

	}
}

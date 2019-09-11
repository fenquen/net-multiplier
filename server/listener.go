package server

import (
	"net"
	"strings"
	"tcp-multiplier/client"
	"tcp-multiplier/config"
)

func ListenAndServe() {
	destTcpSvrAddrStrSlice := strings.Split(config.DestTcpSvrAddrs, config.DELIMITER)

	for {
		localTcpSvrAddr, err := net.ResolveTCPAddr(config.TCP_TYPE, config.LocalTcpSvrAddr)
		if nil != err {
			panic(err)
		}

		tcpListener, err := net.ListenTCP(config.TCP_TYPE, localTcpSvrAddr)

		srcTcpConn, err := tcpListener.AcceptTCP()

		// single srcTcpConn dimension
		go func() {
			senderDataChanSlice := make([]chan []byte, len(destTcpSvrAddrStrSlice), len(destTcpSvrAddrStrSlice))

			for a, destTcpSvrAddrStr := range destTcpSvrAddrStrSlice {
				dataChan := make(chan []byte, 100)
				senderDataChanSlice[a] = dataChan
				go client.Send(dataChan, destTcpSvrAddrStr)
			}

			for {
				tempByteSlice := make([]byte, 1024, 1024)
				_, _ = srcTcpConn.Read(tempByteSlice)

				// need to dispatch data to each sender's data channel
				for _, senderDataChan := range senderDataChanSlice {
					senderDataChan <- tempByteSlice
				}
			}
		}()
	}
}

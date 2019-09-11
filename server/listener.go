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

		// goroutine for single srcTcpConn
		go func() {
			srcDataChanSlice := make([]chan []byte, len(destTcpSvrAddrStrSlice), len(destTcpSvrAddrStrSlice))

			for a, destTcpSvrAddrStr := range destTcpSvrAddrStrSlice {
				srcDataChan := make(chan []byte, 100)
				srcDataChanSlice[a] = srcDataChan

				sender, err := client.NewSender(srcDataChan, destTcpSvrAddrStr)
				if nil != err {
					continue
				}
				sender.Start()
			}

			for {
				tempByteSlice := make([]byte, 1024, 1024)
				readCount, _ := srcTcpConn.Read(tempByteSlice)
				tempByteSlice = tempByteSlice[0:readCount]

				// need to dispatch data to each sender's data channel
				for _, srcDataChan := range srcDataChanSlice {
					srcDataChan <- tempByteSlice
				}
			}
		}()
	}
}

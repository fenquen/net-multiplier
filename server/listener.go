package server

import (
	"io"
	"log"
	"net"
	"strings"
	"tcp-multiplier/client"
	"tcp-multiplier/config"
)

func ListenAndServe() {
	destTcpSvrAddrStrSlice := strings.Split(config.DestTcpSvrAddrs, config.DELIMITER)

	for {
		localTcpSvrAddr, err := net.ResolveTCPAddr(config.TCP_TYPE, config.LocalTcpSvrAddr)

		// fail to bind local address
		if nil != err {
			panic(err)
		}

		tcpListener, err := net.ListenTCP(config.TCP_TYPE, localTcpSvrAddr)

		srcTcpConn, err := tcpListener.AcceptTCP()

		// goroutine for single srcTcpConn
		go func() {
			defer func() {
				_ = srcTcpConn.Close()
			}()

			destNum := len(destTcpSvrAddrStrSlice)

			srcDataChanSlice := make([]chan []byte, destNum, destNum)
			senderSlice := make([]*client.Sender, destNum, destNum)

			for a, destTcpSvrAddrStr := range destTcpSvrAddrStrSlice {
				srcDataChan := make(chan []byte, 100)
				srcDataChanSlice[a] = srcDataChan

				sender, err := client.NewSender(srcDataChan, destTcpSvrAddrStr)

				// fail to build sender,due to net err
				if nil != err {
					srcDataChanSlice[a] = nil
					continue
				}

				senderSlice[a] = sender

				sender.Start()
			}

			// loop
			for {
				tempByteSlice := make([]byte, 1024, 1024)

				readCount, err := srcTcpConn.Read(tempByteSlice)

				// meanings srcTcpConn is closed by client
				if err == io.EOF {
					log.Println("srcTcpConn.Read EOF,srcTcpConn is closed by client")

					// interrupt all sender serving this srcTcpConn
					for _, sender := range senderSlice {
						if nil != sender {
							sender.Interrupt()
						}
					}

					return
				}

				tempByteSlice = tempByteSlice[0:readCount]

				// need to dispatch data to each sender's data channel
				for _, srcDataChan := range srcDataChanSlice {
					if nil == srcDataChan {
						continue
					}
					srcDataChan <- tempByteSlice
				}
			}
		}()
	}
}

package server

import (
	"encoding/hex"
	"io"
	"log"
	"net"
	"strings"
	"tcp-multiplier/client"
	"tcp-multiplier/config"
)

func ListenAndServe() {
	destTcpSvrAddrStrSlice := strings.Split(config.DestSvrAddrs, config.DELIMITER)
	log.Println("destTcpSvrAddr ", destTcpSvrAddrStrSlice)
	destNum := len(destTcpSvrAddrStrSlice)

	localTcpSvrAddr, err := net.ResolveTCPAddr(config.TCP_TYPE, config.LocalSvrAddr)
	if nil != err {
		log.Println("localTcpSvrAddr err")
		panic(err)
	}

	tcpListener, err := net.ListenTCP(config.TCP_TYPE, localTcpSvrAddr)
	if nil != err {
		log.Println("tcpListener err", err)
		panic(err)
	}

	for {
		srcTcpConn, err := tcpListener.AcceptTCP()
		if nil != err {
			log.Println("tcpListener.AcceptTCP() err", err)
			continue
		}

		log.Println("got srcTcpConn", srcTcpConn)

		// goroutine for single srcTcpConn
		go func() {
			defer func() {
				_ = srcTcpConn.Close()
			}()

			var senderSlice []client.Sender
			if destNum > 0 {
				//srcDataChanSlice := make([]chan []byte, destNum, destNum)
				senderSlice = make([]client.Sender, destNum, destNum)

				for a, destTcpSvrAddrStr := range destTcpSvrAddrStrSlice {

					sender, err := client.NewSender(destTcpSvrAddrStr, config.Mode)

					// fail to build sender,due to net err
					if nil != err {
						continue
					}

					senderSlice[a] = sender

					sender.Start()
				}
			}

			// loop
			for {
				tempByteSlice := make([]byte, 1024, 1024)

				readCount, err := srcTcpConn.Read(tempByteSlice)

				// meanings srcTcpConn is closed by client
				if 0 >= readCount && err != io.EOF {
					log.Println("srcTcpConn.Read EOF,srcTcpConn is closed by client")

					if nil != senderSlice {
						// interrupt all sender serving this srcTcpConn
						for _, sender := range senderSlice {
							if nil != sender {
								sender.Interrupt()
							}
						}
					}

					return
				}

				tempByteSlice = tempByteSlice[0:readCount]

				log.Println("receive src data from " + srcTcpConn.RemoteAddr().String())
				log.Println(hex.EncodeToString(tempByteSlice), "\n")

				if nil != senderSlice {
					// need to dispatch data to each sender's data channel
					for _, sender := range senderSlice {
						if nil == senderSlice || sender.IsClosed() {
							continue
						}

						sender.GetSrcDataChan() <- tempByteSlice
					}
				}
			}
		}()
	}
}

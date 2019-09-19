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

var destSvrAddrStrSlice = strings.Split(config.DestSvrAddrs, config.DELIMITER)

func ListenAndServeTcp() {
	log.Println("destSvrAddr ", destSvrAddrStrSlice)

	// mode tcp
	localTcpSvrAddr, err := net.ResolveTCPAddr(config.TCP_MODE, config.LocalSvrAddr)
	if nil != err {
		log.Println("localTcpSvrAddr err")
		panic(err)
	}

	tcpListener, err := net.ListenTCP(config.TCP_MODE, localTcpSvrAddr)
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
		go processConn(srcTcpConn)
	}

}

func ServeUdp() {
	localUdpSvrAddr, err := net.ResolveUDPAddr(config.UDP_MODE, config.LocalSvrAddr)
	if nil != err {
		log.Println("localUdpSvrAddr err")
		panic(err)
	}

	for {
		udpConn, err := net.ListenUDP(config.UDP_MODE, localUdpSvrAddr)
		if nil != err {
			log.Println("ServeUdp net.ListenUDP err ", err)
			panic(err)
		}

		processConn(udpConn)
	}

}

func processConn(srcConn net.Conn) {
	defer func() {
		_ = srcConn.Close()
	}()

	// senderSlice
	var senderSlice []client.Sender
	if nil != destSvrAddrStrSlice && len(destSvrAddrStrSlice) > 0 {
		destSvrNum := len(destSvrAddrStrSlice)
		senderSlice = make([]client.Sender, destSvrNum, destSvrNum)

		for a, destTcpSvrAddrStr := range destSvrAddrStrSlice {
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

		readCount, err := srcConn.Read(tempByteSlice)

		// meanings srcTcpConn is closed by client
		if 0 >= readCount && err != io.EOF {
			log.Println("srcConn.Read(tempByteSlice), 0 >= readCount && err != io.EOF")

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

		log.Println("receive src data from " + srcConn.RemoteAddr().String())
		log.Println(hex.EncodeToString(tempByteSlice), "\n")

		// per dest/sender a goroutine
		go func(senderSlice []client.Sender, data [] byte) {
			if nil == senderSlice {
				return
			}

			// need to dispatch data to each sender's data channel
			for _, sender := range senderSlice {
				if nil == senderSlice || sender.IsClosed() {
					continue
				}

				sender.GetSrcDataChan() <- data
			}
		}(senderSlice, tempByteSlice)
	}
}

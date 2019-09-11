package client

import (
	"log"
	"net"
	"strconv"
	"tcp-multiplier/config"
	"tcp-multiplier/utils"
)

func Send(senderDataChan chan []byte, destTcpSvrAddrStr string) {
	localTcpClientAddr, err := net.ResolveTCPAddr(config.TCP_TYPE,
		config.LocalTcpClientHost+":"+strconv.Itoa(int(utils.GetLocalTcpClientPort())))
	if nil!=err{
		return
	}


	destTcpSvrAddr, _ := net.ResolveTCPAddr(config.TCP_TYPE, destTcpSvrAddrStr)

	tcpConn2Dest, _ := net.DialTCP(config.TCP_TYPE, localTcpClientAddr, destTcpSvrAddr)
	defer func() {
		log.Println(recover())
		_ = tcpConn2Dest.Close()
	}()

	for {
		select {
		case byteSlice := <-senderDataChan:
			_, _ = tcpConn2Dest.Write(byteSlice)
		}
	}
}

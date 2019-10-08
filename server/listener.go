package server

import (
	"encoding/hex"
	"fmt"
	"go.uber.org/zap"
	"net"
	"net-multiplier/client"
	"net-multiplier/config"
	"net-multiplier/zaplog"
	"strings"
)

var destSvrAddrStrSlice = strings.Split(config.DestSvrAddrs, config.DELIMITER)

func ListenAndServeTcp() {
	zaplog.LOGGER.Info("destSvrAddr " + fmt.Sprint(destSvrAddrStrSlice))

	// mode tcp
	localTcpSvrAddr, err := net.ResolveTCPAddr(config.TCP_MODE, config.LocalSvrAddr)
	if nil != err {
		zaplog.LOGGER.Info("localTcpSvrAddr err")
		panic(err)
	}

	tcpListener, err := net.ListenTCP(config.TCP_MODE, localTcpSvrAddr)
	if nil != err {
		zaplog.LOGGER.Info("net.ListenTCP", zap.Any("err", err))
		panic(err)
	}

	for {
		srcTcpConn, err := tcpListener.AcceptTCP()
		if nil != err {
			zaplog.LOGGER.Info("tcpListener.AcceptTCP() err", zap.Any("err", err))
			continue
		}

		zaplog.LOGGER.Info("got srcTcpConn " + fmt.Sprint(srcTcpConn))

		// goroutine for single srcTcpConn
		go processConn(srcTcpConn)
	}

}

func ServeUdp() {
	zaplog.LOGGER.Info("destSvrAddr " + fmt.Sprint(destSvrAddrStrSlice))

	localUdpSvrAddr, err := net.ResolveUDPAddr(config.UDP_MODE, config.LocalSvrAddr)
	if nil != err {
		zaplog.LOGGER.Info("localUdpSvrAddr err")
		panic(err)
	}

	udpConn, err := net.ListenUDP(config.UDP_MODE, localUdpSvrAddr)
	if nil != err {
		zaplog.LOGGER.Info("ServeUdp net.ListenUDP err ", zap.Any("err", err))
		panic(err)
	}

	processConn(udpConn)

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
		tempByteSlice := make([]byte, config.TempByteSliceLen, config.TempByteSliceLen)

		//_ = srcConn.SetReadDeadline(time.Now().Add(time.Second * 10))
		zaplog.LOGGER.Info("before srcConn.Read(tempByteSlice)")
		readCount, err := srcConn.Read(tempByteSlice)
		zaplog.LOGGER.Info("readCount, err := srcConn.Read(tempByteSlice)", zap.Any("readCount", readCount), zap.Any("err", err))

		// meanings srcTcpConn is closed by client
		if 0 >= readCount && err != nil /*io.EOF*/ {
			zaplog.LOGGER.Info("srcConn.Read(tempByteSlice), 0 >= readCount && err != io.EOF")

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

		//zaplog.LOGGER.Info("receive src data from " + srcConn.RemoteAddr().String())
		zaplog.LOGGER.Info(hex.EncodeToString(tempByteSlice))

		// per dest/sender a goroutine
		go func(senderSlice []client.Sender, data [] byte) {
			if nil == senderSlice {
				return
			}

			// need to dispatch data to each sender's data channel
			for _, sender := range senderSlice {
				if sender.IsClosed() {
					continue
				}

				go func(sender client.Sender) {
					sender.GetSrcDataChan() <- data
				}(sender)

			}
		}(senderSlice, tempByteSlice)
	}
}

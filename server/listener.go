package server

import (
	"encoding/hex"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"net"
	"net-multiplier/client"
	"net-multiplier/config"
	"net-multiplier/zaplog"
	"strings"
	"sync"
)

var destSvrAddrStrSlice = strings.Split(*config.DestSvrAddrs, config.DELIMITER)

func ListenAndServeTcp() {
	zaplog.LOGGER.Info("destSvrAddr " + fmt.Sprint(destSvrAddrStrSlice))

	// mode tcp
	localTcpSvrAddr, err := net.ResolveTCPAddr(config.TCP_MODE, *config.LocalSvrAddr)
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
	zaplog.LOGGER.Info("destSvrAddr slice " + fmt.Sprint(destSvrAddrStrSlice))

	localUdpSvrAddr, err := net.ResolveUDPAddr(config.UDP_MODE, *config.LocalSvrAddr)
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

// warm up in advance
func buildAndBootSenderSlice() ([]client.Sender, error) {
	// senderSlice
	var senderSlice []client.Sender

	// actually empty dest slice
	if nil == destSvrAddrStrSlice || len(destSvrAddrStrSlice) == 0 {
		return nil, errors.New("nil == destSvrAddrStrSlice || len(destSvrAddrStrSlice) == 0")
	}

	destSvrNum := len(destSvrAddrStrSlice)
	senderSlice = make([]client.Sender, 0, destSvrNum)

	for _, destTcpSvrAddrStr := range destSvrAddrStrSlice {
		sender, err := client.NewSender(destTcpSvrAddrStr, *config.Mode)

		// fail to build sender,due to net err
		if nil != err {
			zaplog.LOGGER.Error("build sender error whose dest is "+destTcpSvrAddrStr, zap.Any("err", err))
			continue
		}
		senderSlice = append(senderSlice, sender)
		sender.Start()
	}

	if len(senderSlice) == 0 {
		return nil, errors.New("build senders totally failed ")
	}

	return senderSlice, nil
}

func processConn(srcConn net.Conn) {
	defer func() {
		recoveredErr := recover()
		if recoveredErr != nil {
			zaplog.LOGGER.Error("processConn panic ", zap.Any("err", recoveredErr))
		}
		_ = srcConn.Close()
	}()

	senderSlice, err := buildAndBootSenderSlice()
	if nil != err {
		panic(err)
	}

	// loop
	for {
		tempByteSlice := make([]byte, *config.TempByteSliceLen, *config.TempByteSliceLen)

		//_ = srcConn.SetReadDeadline(time.Now().Add(time.Second * 10))
		zaplog.LOGGER.Info("before srcConn.Read(tempByteSlice)")
		readCount, err := srcConn.Read(tempByteSlice)
		zaplog.LOGGER.Info("readCount, err := srcConn.Read(tempByteSlice)",
			zap.Any("readCount", readCount), zap.Any("err", err))

		// meanings srcTcpConn is closed by client
		if 0 >= readCount && err != nil /*io.EOF*/ {
			zaplog.LOGGER.Error("srcConn.Read(tempByteSlice), 0 >= readCount && err != nil",
				zap.Any("err", err))

			if *config.Mode == config.TCP_MODE {
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
		fmt.Println("data received ", hex.EncodeToString(tempByteSlice))

		// per dest/sender a goroutine
		/*go func(senderSlice []client.Sender, tempByteSlice [] byte) {
		if nil == senderSlice {
			return
		}*/

		// need to dispatch data to each sender's data channel
		waitGroup := &sync.WaitGroup{}
		waitGroup.Add(len(senderSlice))
		for _, sender := range senderSlice {
			// sender.interrupt() called by current routine,so current routine can immediately know the state
			if sender == nil {
				waitGroup.Done()
				continue
			}

			select {
			case <-sender.GetReportUnavailableChan():
				sender.Close()
				sender = nil
				waitGroup.Done()
				continue
			default:
			}

			go func(sender client.Sender) {
				sender.GetSrcDataChan() <- tempByteSlice
				waitGroup.Done()
			}(sender)
		}
		waitGroup.Wait()
		//}(senderSlice, tempByteSlice)
	}
}

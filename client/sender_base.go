package client

import (
	"encoding/hex"
	"fmt"
	"go.uber.org/zap"
	"net"
	"net-multiplier/config"
	"net-multiplier/zaplog"
)

type SenderBase struct {
	conn2DestSvr net.Conn
	srcDataChan  chan []byte
	switcher     chan bool
	closed       bool
	localAddr    net.Addr
	remoteAddr   net.Addr
}

func (senderBase *SenderBase) Start() {
	go senderBase.Run()
}

func (senderBase *SenderBase) Run() {
	defer func() {
		recoveredErr := recover()
		if nil != recoveredErr {
			zaplog.LOGGER.Error("recovered error ", zap.Any("err", recoveredErr))
		}
		senderBase.Close();
	}()

	for {
		// whether need to be interrupted
		select {
		case v := <-senderBase.switcher:
			if v {
				return
			}
		default:
		}

		select {
		case byteSlice := <-senderBase.srcDataChan:
			var err error
			switch *config.Mode {
			case config.TCP_MODE:
				_, err = senderBase.conn2DestSvr.Write(byteSlice)
			case config.UDP_MODE:
				udpConn, _ := senderBase.conn2DestSvr.(*net.UDPConn)
				udpAddr, _ := senderBase.remoteAddr.(*net.UDPAddr)
				_, err = udpConn.WriteToUDP(byteSlice, udpAddr)
			}

			if nil != err {
				zaplog.LOGGER.Info("senderBase.conn2DestSvr.Write", zap.Any("err", err))
				return
			}

			zaplog.LOGGER.Info("successfully write data to dest " + hex.EncodeToString(byteSlice))
			zaplog.LOGGER.Info(fmt.Sprint(senderBase.localAddr))
			zaplog.LOGGER.Info(fmt.Sprint(senderBase.remoteAddr))
		}
	}
}

func (senderBase *SenderBase) Interrupt() {
	senderBase.switcher <- true
	//senderBase.Close()
}

func (senderBase *SenderBase) Close() {
	senderBase.closed = true
	_ = senderBase.conn2DestSvr.Close()
	close(senderBase.srcDataChan)
	close(senderBase.switcher)
}

func (senderBase *SenderBase) IsClosed() bool {
	return senderBase.closed
}

func (senderBase *SenderBase) GetSrcDataChan() chan [] byte {
	return senderBase.srcDataChan
}

func (senderBase *SenderBase) SetConn2DestSvr(conn2DestSvr net.Conn) {
	senderBase.conn2DestSvr = conn2DestSvr
}

func (senderBase *SenderBase) SetSrcDataChan(srcDataChan chan []byte) {
	senderBase.srcDataChan = srcDataChan
}
func (senderBase *SenderBase) SetSwitcher(switcher chan bool) {
	senderBase.switcher = switcher
}

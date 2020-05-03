package client

import (
	"encoding/hex"
	"go.uber.org/zap"
	"net"
	"net-multiplier/config"
	"net-multiplier/model"
	"net-multiplier/zaplog"
	"time"
)

type SenderBase struct {
	conn2DestSvr net.Conn
	srcDataChan  chan *model.DataWrapper
	switcher     chan bool
	interrupted  bool
	localAddr    net.Addr
	remoteAddr   net.Addr

	reportUnavailableChan chan bool
	available             bool

	mode string
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

		_ = senderBase.conn2DestSvr.Close()

		// due to panic or interrupt
		senderBase.reportUnavailable()

		//senderBase.Close();
	}()

	for {
		// whether need to be interrupted
		select {
		// interrupted
		case v := <-senderBase.switcher:
			if v {
				return
			}
		default:
		}

		select {
		case dataWrapper, allRight := <-senderBase.srcDataChan:
			// means the chan is closed
			if !allRight {
				return
			}
			var err error
			switch senderBase.mode {
			case config.TCP_MODE:
				_, err = senderBase.conn2DestSvr.Write(dataWrapper.Data)
			case config.UDP_MODE:
				udpConn, _ := senderBase.conn2DestSvr.(*net.UDPConn)
				udpAddr, _ := senderBase.remoteAddr.(*net.UDPAddr)
				_, err = udpConn.WriteToUDP(dataWrapper.Data, udpAddr)
			}

			if nil != err {
				panic(err)
			}

			zaplog.LOGGER.Debug("successfully write data to dest "+hex.EncodeToString(dataWrapper.Data),
				zap.Any("localAddr", "senderBase.localAddr"), zap.Any("remoteAddr", senderBase.remoteAddr))

			dataWrapper.PutBack()
		case <-time.After(time.Millisecond):
		}
	}
}

func (senderBase *SenderBase) reportUnavailable() {
	senderBase.reportUnavailableChan <- true
	senderBase.available = false
	close(senderBase.reportUnavailableChan)
}

func (senderBase *SenderBase) GetReportUnavailableChan() <-chan bool {
	return senderBase.reportUnavailableChan
}

func (senderBase *SenderBase) SetReportUnavailableChan(unavailableChan chan bool) {
	senderBase.reportUnavailableChan = unavailableChan
}

// used by other element
func (senderBase *SenderBase) Interrupt() {
	senderBase.switcher <- true
	senderBase.interrupted = true
	close(senderBase.switcher)
}

// should be triggered by the write side
func (senderBase *SenderBase) Close() {
	close(senderBase.srcDataChan)
}

func (senderBase *SenderBase) Interrupted() bool {
	return senderBase.interrupted
}

func (senderBase *SenderBase) GetSrcDataChan() chan<- *model.DataWrapper {
	return senderBase.srcDataChan
}

func (senderBase *SenderBase) SetConn2DestSvr(conn2DestSvr net.Conn) {
	senderBase.conn2DestSvr = conn2DestSvr
}

func (senderBase *SenderBase) SetSrcDataChan(srcDataChan chan *model.DataWrapper) {
	senderBase.srcDataChan = srcDataChan
}
func (senderBase *SenderBase) SetSwitcher(switcher chan bool) {
	senderBase.switcher = switcher
}

func (senderBase *SenderBase) SetMode(mode string) {
	senderBase.mode = mode
}

func (senderBase *SenderBase) GetMode() string {
	return senderBase.mode
}

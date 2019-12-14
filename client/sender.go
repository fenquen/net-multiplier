package client

import (
	"go.uber.org/zap"
	"net"
	"net-multiplier/config"
	"net-multiplier/utils"
	"net-multiplier/zaplog"
	"strconv"
)

type Sender interface {
	Start()
	Run()

	Interrupt()
	Close()
	Interrupted() bool

	GetSrcDataChan() chan [] byte

	SetConn2DestSvr(conn2DestSvr net.Conn)
	SetSrcDataChan(srcDataChan chan []byte)
	SetSwitcher(switcher chan bool)

	GetReportUnavailableChan() chan bool

	//Write(byteSlice []byte) (int, error)
}

func NewSender(destTcpSvrAddrStr string, mode string) (Sender, error) {
	var conn2DestSvr net.Conn
	var result Sender

	switch mode {
	case config.TCP_MODE:
		// localClientAddr
		localClientAddr, err := net.ResolveTCPAddr(mode,
			*config.LocalClientHost+":"+strconv.Itoa(int(utils.GetLocalTcpClientPort())))
		if nil != err {
			zaplog.LOGGER.Error("ResolveTCPAddr localClientAddr", zap.Any("err", err), zap.Any("LocalClientHost", *config.LocalClientHost))
			return nil, err
		}

		// destSvrAddr
		destSvrAddr, err := net.ResolveTCPAddr(mode, destTcpSvrAddrStr)
		if nil != err {
			zaplog.LOGGER.Error("ResolveTCPAddr destSvrAddr", zap.Any("err", err), zap.Any("destTcpSvrAddrStr", destTcpSvrAddrStr))
			return nil, err
		}

		// conn2DestSvr
		conn2DestSvr, err = net.DialTCP(mode, localClientAddr, destSvrAddr)
		if nil != err {
			zaplog.LOGGER.Error("DialTCP conn2DestSvr", zap.Any("err", err), zap.Any("destTcpSvrAddrStr", destTcpSvrAddrStr))
			return nil, err
		}

		tcpSender := &TcpSender{}
		tcpSender.localAddr = localClientAddr
		tcpSender.remoteAddr = destSvrAddr

		result = tcpSender
	case config.UDP_MODE:
		// localClientAddr
		localClientAddr, err := net.ResolveUDPAddr(mode,
			*config.LocalClientHost+":"+strconv.Itoa(int(utils.GetLocalTcpClientPort())))
		if nil != err {
			zaplog.LOGGER.Error("ResolveUDPAddr localClientAddr", zap.Any("err", err), zap.Any("LocalClientHost", "LocalClientHost"))
			return nil, err
		}

		// destSvrAddr
		destSvrAddr, err := net.ResolveUDPAddr(mode, destTcpSvrAddrStr)
		if nil != err {
			zaplog.LOGGER.Error("ResolveUDPAddr destSvrAddr", zap.Any("err", err), zap.Any("destTcpSvrAddrStr", destTcpSvrAddrStr))
			return nil, err
		}

		// conn2DestSvr
		conn2DestSvr, err = net.ListenUDP(mode, localClientAddr)
		if nil != err {
			zaplog.LOGGER.Error("DialUDP conn2DestSvr", zap.Any("err", err), zap.Any("destTcpSvrAddrStr", destTcpSvrAddrStr))
			return nil, err
		}

		udpSender := &UdpSender{}
		udpSender.localAddr = localClientAddr
		udpSender.remoteAddr = destSvrAddr

		result = udpSender
	}

	result.SetConn2DestSvr(conn2DestSvr)
	result.SetSrcDataChan(make(chan []byte, 100))
	result.SetSwitcher(make(chan bool, 1))

	return result, nil

}

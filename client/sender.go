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

	GetSrcDataChan() chan []byte

	SetConn2DestSvr(conn2DestSvr net.Conn)
	SetSrcDataChan(srcDataChan chan []byte)
	SetSwitcher(switcher chan bool)

	GetReportUnavailableChan() chan bool

	SetMode(mode string)
	GetMode() string

	//Write(byteSlice []byte) (int, error)
}

func NewSender(destTcpSvrAddrStr string, mode string) (Sender, error) {
	var conn2DestSvr net.Conn
	var result Sender

	switch mode {
	case config.TCP_MODE:
		// localClientAddr
		localClientAddr, err := net.ResolveTCPAddr(mode,
			*config.LocalClientHostStr+":"+strconv.Itoa(int(utils.GetLocalClientPort())))
		if nil != err {
			zaplog.LOGGER.Error("ResolveTCPAddr localClientAddr", zap.Any("err", err), zap.Any("LocalClientHostStr", *config.LocalClientHostStr))
			return nil, err
		}

		// destAddr
		destAddr, err := net.ResolveTCPAddr(mode, destTcpSvrAddrStr)
		if nil != err {
			zaplog.LOGGER.Error("ResolveTCPAddr destAddr", zap.Any("err", err), zap.Any("destTcpSvrAddrStr", destTcpSvrAddrStr))
			return nil, err
		}

		// conn2DestSvr
		conn2DestSvr, err = net.DialTCP(mode, localClientAddr, destAddr)
		if nil != err {
			zaplog.LOGGER.Error("DialTCP conn2DestSvr", zap.Any("err", err), zap.Any("destTcpSvrAddrStr", destTcpSvrAddrStr))
			return nil, err
		}

		tcpSender := &TcpSender{}
		tcpSender.localAddr = localClientAddr
		tcpSender.remoteAddr = destAddr

		result = tcpSender
	case config.UDP_MODE:
		// localClientAddr
		localClientAddr, err := net.ResolveUDPAddr(mode,
			*config.LocalClientHostStr+":"+strconv.Itoa(int(utils.GetLocalClientPort())))
		if nil != err {
			zaplog.LOGGER.Error("ResolveUDPAddr localClientAddr", zap.Any("err", err), zap.Any("LocalClientHostStr", "LocalClientHostStr"))
			return nil, err
		}

		// destAddr
		destAddr, err := net.ResolveUDPAddr(mode, destTcpSvrAddrStr)
		if nil != err {
			zaplog.LOGGER.Error("ResolveUDPAddr destAddr", zap.Any("err", err), zap.Any("destTcpSvrAddrStr", destTcpSvrAddrStr))
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
		udpSender.remoteAddr = destAddr

		result = udpSender
	}

	result.SetConn2DestSvr(conn2DestSvr)
	result.SetSrcDataChan(make(chan []byte, 100))
	result.SetSwitcher(make(chan bool, 1))

	return result, nil

}

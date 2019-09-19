package client

import (
	"log"
	"net"
	"strconv"
	"tcp-multiplier/config"
	"tcp-multiplier/utils"
)


func NewTcpSender( destTcpSvrAddrStr string) (*TcpSender, error) {
	// localTcpClientAddr
	localTcpClientAddr, err := net.ResolveTCPAddr(config.TCP_TYPE,
		config.LocalTcpClientHost+":"+strconv.Itoa(int(utils.GetLocalTcpClientPort())))
	if nil != err {
		log.Println("ResolveTCPAddr localTcpClientAddr  err", err, destTcpSvrAddrStr)
		return nil, err
	}

	// destTcpSvrAddr
	destTcpSvrAddr, err := net.ResolveTCPAddr(config.TCP_TYPE, destTcpSvrAddrStr)
	if nil != err {
		log.Println("ResolveTCPAddr destTcpSvrAddr err", err, destTcpSvrAddrStr)
		return nil, err
	}

	// tcpConn2DestSvr
	tcpConn2DestSvr, err := net.DialTCP(config.TCP_TYPE, localTcpClientAddr, destTcpSvrAddr)
	if nil != err {
		log.Println("DialTCP tcpConn2DestSvr err", err, destTcpSvrAddrStr)
		return nil, err
	}

	sender := &TcpSender{}
	sender.tcpConn2DestSvr = tcpConn2DestSvr
	sender.srcDataChan = make(chan []byte, 100)
	sender.switcher = make(chan bool, 1)

	return sender, nil
}

type TcpSender struct {
	tcpConn2DestSvr *net.TCPConn
	srcDataChan     chan []byte
	switcher        chan bool
	closed          bool
}

func (sender *TcpSender) Start() {
	go sender.Run()
}

func (sender *TcpSender) Run() {
	defer func() {
		recover()

		sender.Close();
	}()

	for {
		// whether need to be interrupted
		select {
		case v := <-sender.switcher:
			if v {
				return
			}
		default:
		}

		select {
		case byteSlice := <-sender.srcDataChan:
			_, err := sender.tcpConn2DestSvr.Write(byteSlice)
			if nil != err {
				log.Println("sender.tcpConn2DestSvr.Write err ", err)
				return
			}
		}
	}
}

func (sender *TcpSender) Interrupt() {
	sender.switcher <- true
	sender.Close()
}

func (sender *TcpSender) Close() {
	sender.closed = true
	_ = sender.tcpConn2DestSvr.Close()
	close(sender.srcDataChan)
	close(sender.switcher)
}

func (sender *TcpSender) IsClosed() bool {
	return sender.closed
}

func (sender *TcpSender) GetSrcDataChan() chan [] byte {

	return sender.srcDataChan
}

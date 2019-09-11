package client

import (
	"log"
	"net"
	"strconv"
	"tcp-multiplier/config"
	"tcp-multiplier/utils"
)

func NewSender(srcDataChan chan []byte, destTcpSvrAddrStr string) (*Sender, error) {
	// local addr
	localTcpClientAddr, err := net.ResolveTCPAddr(config.TCP_TYPE,
		config.LocalTcpClientHost+":"+strconv.Itoa(int(utils.GetLocalTcpClientPort())))
	if nil != err {
		log.Println("ResolveTCPAddr localTcpClientAddr  err", err, destTcpSvrAddrStr)
		return nil, err
	}

	// dest addr
	destTcpSvrAddr, err := net.ResolveTCPAddr(config.TCP_TYPE, destTcpSvrAddrStr)
	if nil != err {
		log.Println("ResolveTCPAddr destTcpSvrAddr err", err, destTcpSvrAddrStr)
		return nil, err
	}

	// tcpConn
	tcpConn2DestSvr, err := net.DialTCP(config.TCP_TYPE, localTcpClientAddr, destTcpSvrAddr)
	if nil != err {
		log.Println("DialTCP tcpConn2DestSvr err", err, destTcpSvrAddrStr)
		return nil, err
	}

	sender := &Sender{}
	sender.tcpConn2DestSvr = tcpConn2DestSvr
	sender.srcDataChan = srcDataChan
	sender.switcher = make(chan bool, 1)

	return sender, nil
}

type Sender struct {
	tcpConn2DestSvr *net.TCPConn
	srcDataChan     chan []byte
	switcher        chan bool
	closed          bool
}

func (sender *Sender) Start() {
	go sender.run()
}

func (sender *Sender) run() {
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

func (sender *Sender) Interrupt() {
	sender.switcher <- true
	sender.Close()
}

func (sender *Sender) Close() {
	sender.closed = true
	_ = sender.tcpConn2DestSvr.Close()
	close(sender.srcDataChan)
	close(sender.switcher)
}

func (sender *Sender) IsClosed() bool {
	return sender.closed
}

func (sender *Sender) GetSrcDataChan() chan [] byte {

	return sender.srcDataChan
}

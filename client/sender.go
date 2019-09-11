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
		log.Println("ResolveTCPAddr localTcpClientAddr  err", err)
		return nil, err
	}

	// dest addr
	destTcpSvrAddr, err := net.ResolveTCPAddr(config.TCP_TYPE, destTcpSvrAddrStr)
	if nil != err {
		log.Println("ResolveTCPAddr destTcpSvrAddr err", err)
		return nil, err
	}

	// tcpConn
	tcpConn2DestSvr, err := net.DialTCP(config.TCP_TYPE, localTcpClientAddr, destTcpSvrAddr)
	if nil != err {
		log.Println("DialTCP tcpConn2DestSvr err", err)
		return nil, err
	}

	sender := &Sender{}
	sender.tcpConn2DestSvr = tcpConn2DestSvr
	sender.srcDataChan = srcDataChan

	return sender, nil
}

type Sender struct {
	tcpConn2DestSvr *net.TCPConn
	srcDataChan     chan []byte
	switcher        chan bool
}

func (sender *Sender) Start() {
	go sender.run()
}

func (sender *Sender) run() {
	defer func() {
		log.Println(recover())

		_ = sender.tcpConn2DestSvr.Close()
		close(sender.srcDataChan)
		close(sender.switcher)
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
			_, _ = sender.tcpConn2DestSvr.Write(byteSlice)
		}
	}
}

func (sender *Sender) Interrupt() {
	sender.switcher <- true
}

package client

import (
	"log"
	"net"
	"strconv"
	"tcp-multiplier/config"
	"tcp-multiplier/utils"
)

func Send(senderDataChan chan []byte, destTcpSvrAddrStr string) {
	localTcpClientAddr, _ := net.ResolveTCPAddr(config.TCP_TYPE,
		config.LocalTcpClientHost+":"+strconv.Itoa(int(utils.GetLocalTcpClientPort())))
	destTcpSvrAddr, _ := net.ResolveTCPAddr(config.TCP_TYPE, destTcpSvrAddrStr)

	tcpConn2DestSvr, _ := net.DialTCP(config.TCP_TYPE, localTcpClientAddr, destTcpSvrAddr)
	defer func() {
		log.Println(recover())
		_ = tcpConn2DestSvr.Close()
	}()

	for {
		select {
		case byteSlice := <-senderDataChan:
			_, _ = tcpConn2DestSvr.Write(byteSlice)
		}
	}
}

type Sender struct {
	tcpConn2DestSvr *net.TCPConn
	srcDataChan     chan []byte
	switcher        chan bool
}

func NewSender(srcDataChan chan []byte, destTcpSvrAddrStr string) (*Sender, error) {
	localTcpClientAddr, err := net.ResolveTCPAddr(config.TCP_TYPE,
		config.LocalTcpClientHost+":"+strconv.Itoa(int(utils.GetLocalTcpClientPort())))

	destTcpSvrAddr, err := net.ResolveTCPAddr(config.TCP_TYPE, destTcpSvrAddrStr)

	tcpConn2DestSvr, err := net.DialTCP(config.TCP_TYPE, localTcpClientAddr, destTcpSvrAddr)

	if nil != err {
		log.Println("NewSender err", err)
		return nil, err
	}

	sender := &Sender{}
	sender.tcpConn2DestSvr = tcpConn2DestSvr
	sender.srcDataChan = srcDataChan

	return sender, nil
}

func (sender *Sender) Start() {
	go sender.run()
}

func (sender *Sender) run() {
	defer func() {
		log.Println(recover())
		_ = sender.tcpConn2DestSvr.Close()
		close(sender.srcDataChan)
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

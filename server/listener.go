package server

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"net"
	"net-multiplier/client"
	"net-multiplier/config"
	"net-multiplier/model/httpResponse"
	"net-multiplier/zaplog"
	"net/http"
	"strings"
	"sync"
)

var destAddrStrSlice = strings.Split(*config.DestSvrAddrs, config.DELIMITER)
var destAddr_sender = make(map[string]client.Sender, 8)
var mutex sync.Mutex

func ListenAndServeTcp() {
	zaplog.LOGGER.Info("destSvrAddr " + fmt.Sprint(destAddrStrSlice))

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
	zaplog.LOGGER.Info("destSvrAddr slice " + fmt.Sprint(destAddrStrSlice))

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
func buildAndBootSenderSlice() error {
	// senderSlice

	// actually empty dest slice
	if nil == destAddrStrSlice || len(destAddrStrSlice) == 0 {
		return errors.New("nil == destAddrStrSlice || len(destAddrStrSlice) == 0")
	}

	// := len(destAddrStrSlice)
	//senderSlice := make([]client.Sender, 0, destSvrNum)

	for _, destAddrStr := range destAddrStrSlice {
		sender, err := client.NewSender(destAddrStr, *config.Mode)

		// fail to build sender,due to net err
		if nil != err {
			zaplog.LOGGER.Error("build sender error whose dest is "+destAddrStr, zap.Any("err", err))
			continue
		}

		destAddr_sender[destAddrStr] = sender
		//senderSlice = append(senderSlice, sender)
		sender.Start()
	}

	if len(destAddr_sender) == 0 {
		return errors.New("build senders totally failed ")
	}

	return nil
}

func processConn(srcConn net.Conn) {
	defer func() {
		recoveredErr := recover()
		if recoveredErr != nil {
			zaplog.LOGGER.Error("processConn panic ", zap.Any("err", recoveredErr))
		}
		_ = srcConn.Close()
	}()

	err := buildAndBootSenderSlice()
	if nil != err {
		panic(err)
	}

	// loop
	for {
		tempByteSlice := make([]byte, *config.TempByteSliceLen, *config.TempByteSliceLen)

		//_ = srcConn.SetReadDeadline(time.Now().Add(time.Second * 10))
		zaplog.LOGGER.Debug("before srcConn.Read(tempByteSlice)")
		readCount, err := srcConn.Read(tempByteSlice)
		zaplog.LOGGER.Debug("readCount, err := srcConn.Read(tempByteSlice)",
			zap.Any("readCount", readCount), zap.Any("err", err))

		// meanings srcTcpConn is closed by client
		if 0 >= readCount && err != nil /*io.EOF*/ {
			zaplog.LOGGER.Error("srcConn.Read(tempByteSlice), 0 >= readCount && err != nil",
				zap.Any("err", err))

			if *config.Mode == config.TCP_MODE {
				// interrupt all sender serving this srcTcpConn
				for _, sender := range destAddr_sender {
					if nil != sender {
						sender.Interrupt()
						sender.Close()
					}
				}
			}

			return
		}

		tempByteSlice = tempByteSlice[0:readCount]

		//zaplog.LOGGER.Info("receive src data from " + srcConn.RemoteAddr().String())
		zaplog.LOGGER.Debug("data received " + hex.EncodeToString(tempByteSlice))

		// per dest/sender a goroutine
		/*go func(senderSlice []client.Sender, tempByteSlice [] byte) {
		if nil == senderSlice {
			return
		}*/

		// need to dispatch data to each sender's data channel
		waitGroup := &sync.WaitGroup{}

		mutex.Lock()
		waitGroup.Add(len(destAddr_sender))
		for _, sender := range destAddr_sender {
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
		mutex.Unlock()

		waitGroup.Wait()
		//}(senderSlice, tempByteSlice)
	}
}

func ServeHttp() {

	handlePanic := func(writer http.ResponseWriter) {
		recoverErr := recover()
		if recoverErr == nil {
			return
		}

		zaplog.LOGGER.Error("", zap.Any("recoverErr", recoverErr))

		response := httpResponse.Fail(fmt.Sprint(recoverErr))
		byteSlice, _ := json.Marshal(response)
		_, _ = writer.Write(byteSlice)
	}

	http.HandleFunc("/multiplier/addDests", func(writer http.ResponseWriter, request *http.Request) {
		defer handlePanic(writer)

		destAddrStrs := request.FormValue("destAddrStrs")

		mutex.Lock()
		for _, destAddrStr := range strings.Split(destAddrStrs, ",") {

			sender, err := client.NewSender(destAddrStr, *config.Mode)

			// fail to build sender,due to net err
			if nil != err {
				zaplog.LOGGER.Error("build sender error whose dest is "+destAddrStr, zap.Any("err", err))
				continue
			}

			destAddr_sender[destAddrStr] = sender

			sender.Start()
		}
		mutex.Unlock()

		_, _ = writer.Write(httpResponse.SUCCESS)
	})

	http.HandleFunc("/multiplier/delDests", func(writer http.ResponseWriter, request *http.Request) {
		defer handlePanic(writer)

		destAddrStrs := request.FormValue("destAddrStrs")

		mutex.Lock()
		for _, destAddrStr := range strings.Split(destAddrStrs, ",") {
			sender := destAddr_sender[destAddrStr]
			if sender == nil {
				continue
			}

			delete(destAddr_sender, destAddrStr)

			sender.Interrupt()
		}
		mutex.Unlock()

		_, _ = writer.Write(httpResponse.SUCCESS)
	})

	if err := http.ListenAndServe(*config.LocalHttpSvrAddr, nil); err != nil {
		zaplog.LOGGER.Info("http.ListenAndServe err")
		panic(err)
	}
}
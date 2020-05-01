package server

import (
	"encoding/hex"
	"encoding/json"
	_ "errors"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"net"
	"net-multiplier/client"
	"net-multiplier/config"
	"net-multiplier/model"
	"net-multiplier/utils"
	"net-multiplier/zaplog"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

//var destAddrStrSlice = strings.Split(*config.DestSvrAddrs, config.DELIMITER)
//var destAddr_sender = make(map[string]client.Sender, 8)
var mutex sync.Mutex
var uuid_task = make(map[string]*model.Task)

func ServeHttp() {

	handlePanic := func(writer http.ResponseWriter) {
		recoverErr := recover()
		if recoverErr == nil {
			return
		}

		zaplog.LOGGER.Error("", zap.Any("recoverErr", recoverErr))

		response := model.Fail(fmt.Sprint(recoverErr))
		byteSlice, _ := json.Marshal(response)
		_, _ = writer.Write(byteSlice)
	}

	http.HandleFunc("/multiplier/addTask", func(writer http.ResponseWriter, request *http.Request) {
		defer handlePanic(writer)

		destAddrsStr := request.FormValue("destAddrsStr")
		mode := request.FormValue("mode")

		mutex.Lock()

		// build senders
		senderSlice, err := buildSenders(destAddrsStr, mode)
		if err != nil {
			for _, sender := range senderSlice {
				sender.Interrupt()
			}
			panic(err)
		}

		// build listener
		err, task := buildLocalSvr(mode, senderSlice)
		if err != nil {
			for _, sender := range senderSlice {
				sender.Interrupt()
			}
			panic(err)
		}

		uuid_task[task.Id] = task

		mutex.Unlock()

		byteSlice, _ := json.Marshal(model.Success(task))
		_, _ = writer.Write(byteSlice)
	})

	http.HandleFunc("/multiplier/delTask", func(writer http.ResponseWriter, request *http.Request) {
		defer handlePanic(writer)

		taskId := request.FormValue("taskId")

		mutex.Lock()
		task := uuid_task[taskId]
		if nil != task {
			delete(uuid_task, taskId)
			_ = task.Close()
		}
		mutex.Unlock()

		_, _ = writer.Write(model.SUCCESS)
	})

	if err := http.ListenAndServe(*config.LocalHttpSvrAddr, nil); err != nil {
		zaplog.LOGGER.Info("http.ListenAndServe err")
		panic(err)
	}
}

func buildSenders(destAddrsStr, mode string) ([]client.Sender, error) {
	destAddrStrSlice := strings.Split(destAddrsStr, config.DELIMITER)
	senderSlice := make([]client.Sender, len(destAddrStrSlice))

	for _, destAddrStr := range destAddrStrSlice {

		sender, err := client.NewSender(destAddrStr, mode)

		// fail to build sender,due to net err
		if nil != err {
			zaplog.LOGGER.Error("build sender error whose dest is "+destAddrStr, zap.Any("err", err))
			continue
		}

		senderSlice = append(senderSlice, sender)
		//destAddr_sender[destAddrStr] = sender

		sender.Start()
	}

	return senderSlice, nil
}

func buildLocalSvr(mode string, senderSlice []client.Sender) (error, *model.Task) {
	// use default mode configured in config
	if mode == "" {
		mode = *config.DefaultMode
	}

	localClientPort := utils.GetLocalClientPort()
	localSvrAddrStr := *config.LocalSvrHostStr + ":" + strconv.Itoa(int(localClientPort))

	var err error
	var task *model.Task

	switch mode {
	case config.TCP_MODE:
		err, task = listenAndServeTcp(localSvrAddrStr, senderSlice)
	case config.UDP_MODE:
		err, task = serveUdp(localSvrAddrStr, senderSlice)
	}

	if err != nil {
		return err, nil
	}

	return nil, task
}

func listenAndServeTcp(localTcpSvrAddrStr string, senderSlice []client.Sender) (error, *model.Task) {
	// zaplog.LOGGER.Info("destSvrAddr " + fmt.Sprint(destAddrStrSlice))

	// mode tcp
	localTcpSvrAddr, err := net.ResolveTCPAddr(config.TCP_MODE, localTcpSvrAddrStr)
	if nil != err {
		zaplog.LOGGER.Info("localTcpSvrAddr err")
		return err, nil
	}

	tcpListener, err := net.ListenTCP(config.TCP_MODE, localTcpSvrAddr)
	if nil != err {
		zaplog.LOGGER.Info("net.ListenTCP", zap.Any("err", err))
		return err, nil
	}

	// build task
	uuidStr := uuid.NewV1().String()
	task := &model.Task{}
	task.Id = uuidStr
	task.LocalSvrAddrStr = localTcpSvrAddrStr
	task.SenderSlice = senderSlice
	task.LocalServer = tcpListener
	task.Mode = config.TCP_MODE

	go func() {
		for {
			srcTcpConn, err := tcpListener.AcceptTCP()
			if nil != err {
				zaplog.LOGGER.Info("tcpListener.AcceptTCP() err", zap.Any("err", err))
				continue
			}

			zaplog.LOGGER.Info("got srcTcpConn " + fmt.Sprint(srcTcpConn))

			// goroutine for single srcTcpConn
			go processConn(srcTcpConn, senderSlice, task)
		}
	}()

	return nil, task
}

func serveUdp(localUdpSvrAddrStr string, senderSlice []client.Sender) (error, *model.Task) {
	//zaplog.LOGGER.Info("destSvrAddr slice " + fmt.Sprint(destAddrStrSlice))

	localUdpSvrAddr, err := net.ResolveUDPAddr(config.UDP_MODE, localUdpSvrAddrStr)
	if nil != err {
		zaplog.LOGGER.Info("localUdpSvrAddr err")
		return err, nil
	}

	udpConn, err := net.ListenUDP(config.UDP_MODE, localUdpSvrAddr)
	if nil != err {
		zaplog.LOGGER.Info("serveUdp net.ListenUDP err ", zap.Any("err", err))
		return err, nil
	}

	// build task
	uuidStr := uuid.NewV1().String()
	task := &model.Task{}
	task.Id = uuidStr
	task.LocalSvrAddrStr = localUdpSvrAddrStr
	task.SenderSlice = senderSlice
	task.LocalServer = udpConn
	task.Mode = config.UDP_MODE

	go processConn(udpConn, senderSlice, task)

	return nil, task

}

// warm up in advance
/*func buildAndBootSenderSlice() error {
	// senderSlice

	// actually empty dest slice
	if nil == destAddrStrSlice || len(destAddrStrSlice) == 0 {
		return errors.New("nil == destAddrStrSlice || len(destAddrStrSlice) == 0")
	}

	// := len(destAddrStrSlice)
	//senderSlice := make([]client.Sender, 0, destSvrNum)

	for _, destAddrStr := range destAddrStrSlice {
		sender, err := client.NewSender(destAddrStr, *config.DefaultMode)

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
}*/

func processConn(srcConn net.Conn, senderSlice []client.Sender, task *model.Task) {
	defer func() {
		recoveredErr := recover()
		if recoveredErr != nil {
			zaplog.LOGGER.Error("processConn panic ", zap.Any("err", recoveredErr))
		}
		_ = srcConn.Close()
	}()

	/*err := buildAndBootSenderSlice()
	if nil != err {
		panic(err)
	}*/

	// loop
	for {
		tempByteSlice := make([]byte, *config.TempByteSliceLen, *config.TempByteSliceLen)

		//_ = srcConn.SetReadDeadline(time.Now().Add(time.Second * 10))
		zaplog.LOGGER.Debug("before srcConn.Read(tempByteSlice)")
		readCount, err := srcConn.Read(tempByteSlice)
		zaplog.LOGGER.Debug("readCount, err := srcConn.Read(tempByteSlice)",
			zap.Any("readCount", readCount), zap.Any("err", err))

		if 0 >= readCount && err != nil /*io.EOF*/ {
			zaplog.LOGGER.Error("srcConn.Read(tempByteSlice), 0 >= readCount && err != nil",
				zap.Any("err", err))

			// meanings srcTcpConn is closed by client
			if task.Mode == config.TCP_MODE {
				// interrupt all sender serving this srcTcpConn
				/*	for _, sender := range senderSlice {
					if nil != sender {
						sender.Interrupt()
						sender.Close()
					}
				}*/

				_ = task.Close()
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
		waitGroup.Add(len(senderSlice))
		for _, sender := range senderSlice {
			// sender.interrupt() called by current routine,so current routine can immediately know the state
			if sender == nil {
				waitGroup.Done()
				continue
			}

			select {
			case <-sender.GetReportUnavailableChan():
				// close the stcDataChan at the write side
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

package server

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	_ "errors"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"io"
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
	"time"
)

var mutex sync.Mutex
var uuid_task = make(map[string]*client.Task)

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
		tempByteSliceLenStr := request.FormValue("tempByteSliceLen")

		// handle tempByteSliceLenStr
		tempByteSliceLen, err := strconv.Atoi(tempByteSliceLenStr)
		if err != nil {
			zaplog.LOGGER.Info("tempByteSliceLenStr can not be converted to int", zap.Any("tempByteSliceLenStr", tempByteSliceLenStr))
			tempByteSliceLen = *config.DefaultTempByteSliceLen
		}
		if tempByteSliceLen == 0 {
			tempByteSliceLen = *config.DefaultTempByteSliceLen
		}

		// build senders
		senderSlice, err := buildSenders(destAddrsStr, mode)
		if err != nil {
			for _, sender := range senderSlice {
				sender.Cancel()
			}
			panic(err)
		}

		// build listener
		err, task := buildLocalSvr(mode, senderSlice, tempByteSliceLen)
		if err != nil {
			for _, sender := range senderSlice {
				sender.Cancel()
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
			task.Cancel()
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

	if len(senderSlice) == 0 {
		return nil, errors.New("haven't successfully build a sender")
	}

	return senderSlice, nil
}

func buildLocalSvr(mode string, senderSlice []client.Sender, tempByteSliceLen int) (error, *client.Task) {
	// use default mode configured in config
	if mode == "" {
		mode = *config.DefaultMode
	}

	localClientPort := utils.GetLocalClientPort()
	localSvrAddrStr := *config.LocalSvrHostStr + ":" + strconv.Itoa(int(localClientPort))

	var err error
	var task *client.Task

	switch mode {
	case config.TCP_MODE:
		err, task = listenAndServeTcp(localSvrAddrStr, senderSlice, tempByteSliceLen)
	case config.UDP_MODE:
		err, task = serveUdp(localSvrAddrStr, senderSlice, tempByteSliceLen)
	}

	if err != nil {
		return err, nil
	}

	return nil, task
}

func listenAndServeTcp(localTcpSvrAddrStr string, senderSlice []client.Sender, tempByteSliceLen int) (error, *client.Task) {
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
	task := BuildTask(localTcpSvrAddrStr, senderSlice, tcpListener, tempByteSliceLen, config.TCP_MODE)

	go func() {
		for {
			srcTcpConn, err := tcpListener.AcceptTCP()
			if nil != err {
				zaplog.LOGGER.Info("tcpListener.AcceptTCP() err", zap.Any("err", err))
				continue
			}

			zaplog.LOGGER.Info("got srcTcpConn " + fmt.Sprint(srcTcpConn))

			// goroutine for single srcTcpConn
			go processConn(srcTcpConn, task)
		}
	}()

	return nil, task
}

func serveUdp(localUdpSvrAddrStr string, senderSlice []client.Sender, tempByteSliceLen int) (error, *client.Task) {
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
	task := BuildTask(localUdpSvrAddrStr, senderSlice, udpConn, tempByteSliceLen, config.UDP_MODE)

	go processConn(udpConn, task)

	return nil, task
}

func BuildTask(localTcpSvrAddrStr string, senderSlice []client.Sender, localServer io.Closer, tempByteSliceLen int, mode string) *client.Task {
	task := &client.Task{}
	task.Id = uuid.NewV1().String()
	task.LocalSvrAddrStr = localTcpSvrAddrStr
	task.SenderSlice = senderSlice
	task.LocalServer = localServer
	task.Mode = mode
	task.TempByteSliceLen = tempByteSliceLen
	task.DataBufWrapperChan = make(chan *client.DataBufWrapper, 1024)
	task.CancelSignalChan = make(chan bool, 1)
	return task
}

func processConn(srcConn net.Conn, task *client.Task) {
	defer func() {
		recoveredErr := recover()
		if recoveredErr != nil {
			zaplog.LOGGER.Error("processConn panic ", zap.Any("err", recoveredErr))
		}
		_ = task.Close()
	}()

	// loop
	for {
		// always make sure the task is valid first
		select {
		// means the task is not valid
		case <-task.CancelSignalChan:
			return
		default:
		}

		var dataWrapper *client.DataBufWrapper

		// try to reuse dataWrapper
		select {
		case d := <-task.DataBufWrapperChan:
			dataWrapper = d
		default:
			dataWrapper = client.BuildDataBufWrapper(task.TempByteSliceLen, int32(len(task.SenderSlice)))
		}

		//tempByteSlice := make([]byte, task.TempByteSliceLen, task.TempByteSliceLen)

		_ = srcConn.SetReadDeadline(time.Now().Add(time.Second * 10))
		zaplog.LOGGER.Debug("before srcConn.Read(tempByteSlice)")
		readCount, err := srcConn.Read(dataWrapper.DataBuf)
		zaplog.LOGGER.Debug("readCount, err := srcConn.Read(tempByteSlice)",
			zap.Any("readCount", readCount), zap.Any("err", err))

		if 0 >= readCount && err != nil /*io.EOF*/ {
			zaplog.LOGGER.Error("srcConn.Read(tempByteSlice), 0 >= readCount && err != nil",
				zap.Any("err", err))

			if netErr, ok := err.(net.Error); ok {
				if netErr.Timeout() || netErr.Temporary() {
					continue
				}
			}

			// meanings srcTcpConn is closed by client
			if task.Mode == config.TCP_MODE {
				// interrupt all sender serving this srcTcpConn
				/*	for _, sender := range senderSlice {
					if nil != sender {
						sender.Cancel()
						sender.Close()
					}
				}*/
			}

			return
		}

		dataWrapper.DataBuf = dataWrapper.DataBuf[0:readCount]

		//zaplog.LOGGER.Info("receive src data from " + srcConn.RemoteAddr().String())
		zaplog.LOGGER.Debug("data received " + hex.EncodeToString(dataWrapper.DataBuf))

		// need to dispatch data to each sender's data channel
		waitGroup := &sync.WaitGroup{}

		mutex.Lock()
		waitGroup.Add(len(task.SenderSlice))
		for _, sender := range task.SenderSlice {
			if sender == nil {
				dataWrapper.PutBack()
				waitGroup.Done()
				continue
			}

			select {
			case <-sender.GetReportUnavailableChan():
				// close the stcDataChan at the write side
				sender.Close()
				sender = nil
				waitGroup.Done()
				dataWrapper.PutBack()
				continue
			default:
			}

			go func(sender client.Sender) {
				sender.GetSrcDataChan() <- dataWrapper
				waitGroup.Done()
			}(sender)
		}
		mutex.Unlock()

		waitGroup.Wait()
		//}(senderSlice, tempByteSlice)
	}
}

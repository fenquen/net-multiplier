package model

import (
	uuid "github.com/satori/go.uuid"
	"io"
	"net-multiplier/client"
)

// the model to describe a single task
type Task struct {
	Id                 string               `json:"id"`
	LocalSvrAddrStr    string               `json:"destAddr"`
	LocalServer        io.Closer            `json:"-"`
	SenderSlice        []client.Sender      `json:"-"`
	Mode               string               `json:"-"`
	TempByteSliceLen   int                  `json:"-"`
	DataBufWrapperChan chan *DataBufWrapper `json:"-"`
	CancelSignalChan   chan bool            `json:"-"`
}

func (task *Task) Close() error {

	_ = task.LocalServer.Close()

	for _, sender := range task.SenderSlice {
		sender.Cancel()
	}

	return nil
}

func (task *Task) Cancel() {
	task.CancelSignalChan <- true
}

func BuildTask(localTcpSvrAddrStr string, senderSlice []client.Sender, localServer io.Closer, tempByteSliceLen int, mode string) *Task {
	task := &Task{}
	task.Id = uuid.NewV1().String()
	task.LocalSvrAddrStr = localTcpSvrAddrStr
	task.SenderSlice = senderSlice
	task.LocalServer = localServer
	task.Mode = mode
	task.TempByteSliceLen = tempByteSliceLen
	task.DataBufWrapperChan = make(chan *DataBufWrapper, 1024)
	task.CancelSignalChan = make(chan bool, 1)
	return task
}

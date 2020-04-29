package model

import (
	"io"
	"net-multiplier/client"
)

// the model to describe a single task
type Task struct {
	Id              string          `json:"id"`
	LocalSvrAddrStr string          `json:"destAddr"`
	LocalServer     io.Closer       `json:"-"`
	SenderSlice     []client.Sender `json:"-"`
	Mode            string          `json:"mode"`
}

func (task *Task) Close() error {

	_ = task.LocalServer.Close()

	for _, sender := range task.SenderSlice {
		sender.Interrupt()
	}

	return nil
}
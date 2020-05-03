package model

import (
	"net-multiplier/server"
	"sync/atomic"
)

type DataWrapper struct {
	Data    []byte
	num     int32
	counter int32
}

func NewDataWrapper(tempSize int, num int32) *DataWrapper {
	dataWrapper := &DataWrapper{}

	dataWrapper.num = num
	dataWrapper.Data = make([]byte, tempSize, tempSize)
	dataWrapper.counter = 0

	return dataWrapper
}

func (dataWrapper *DataWrapper) PutBack() {
	if atomic.AddInt32(&dataWrapper.counter, 1) == dataWrapper.num {
		select {
		case server.DataWrapperChan <- dataWrapper:
		default:

		}
	}
}

package client

import (
	"sync/atomic"
)

type DataBufWrapper struct {
	DataBuf       []byte
	num           int32
	counter       int32
	tempSize      int
	BelongingTask *Task
}

func BuildDataBufWrapper(tempSize int, num int32) *DataBufWrapper {
	dataWrapper := &DataBufWrapper{}

	dataWrapper.num = num
	dataWrapper.DataBuf = make([]byte, tempSize, tempSize)
	dataWrapper.counter = 0
	dataWrapper.tempSize = tempSize

	return dataWrapper
}

func (dataWrapper *DataBufWrapper) PutBack() {
	if atomic.AddInt32(&dataWrapper.counter, 1) == dataWrapper.num {
		select {
		case dataWrapper.BelongingTask.DataBufWrapperChan <- dataWrapper:
			// reset the byte slice len and cap
			dataWrapper.DataBuf = dataWrapper.DataBuf[0:dataWrapper.tempSize]
			atomic.StoreInt32(&dataWrapper.counter, 0)
		default:
			// discard this dataWrapper
		}
	}
}

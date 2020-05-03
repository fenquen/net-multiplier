package model

import (
	"net-multiplier/server"
	"sync/atomic"
)

type DataWrapper struct {
	Data     []byte
	num      int32
	counter  int32
	tempSize int
}

func NewDataWrapper(tempSize int, num int32) *DataWrapper {
	dataWrapper := &DataWrapper{}

	dataWrapper.num = num
	dataWrapper.Data = make([]byte, tempSize, tempSize)
	dataWrapper.counter = 0
	dataWrapper.tempSize = tempSize

	return dataWrapper
}

func (dataWrapper *DataWrapper) PutBack() {
	if atomic.AddInt32(&dataWrapper.counter, 1) == dataWrapper.num {
		select {
		case server.DataWrapperChan <- dataWrapper:
			// reset the byte slice len and cap
			dataWrapper.Data = dataWrapper.Data[0:dataWrapper.tempSize]
		default:
			// discard this dataWrapper
		}
	}
}

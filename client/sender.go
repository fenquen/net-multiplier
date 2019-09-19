package client
type Sender interface {
	Start()
	Run()
	Interrupt()
	Close()
	IsClosed() bool
	GetSrcDataChan() chan [] byte
}
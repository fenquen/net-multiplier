package main

import (
	"fmt"
	"tcp-multiplier/config"
)

func main() {
	fmt.Println(config.LocalTcpSvrAddr)
	fmt.Println(config.DestTcpSvrAddr)
}

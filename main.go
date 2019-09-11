package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"tcp-multiplier/config"
)

func main() {
	fmt.Println(config.LocalTcpSvrAddr)
	fmt.Println(config.DestTcpSvrAddrs)

	if "" == strings.Trim(config.DestTcpSvrAddrs, " ") {
		log.Println("you actually did not specify a valid \"DestTcpSvrAddrs\",it is virtually empty")
		flag.Usage()
		os.Exit(0)
	}
}

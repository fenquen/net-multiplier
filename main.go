package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"tcp-multiplier/config"
)

func main0() {
	if "" == strings.Trim(config.DestTcpSvrAddrs, " ") {
		log.Println("you actually did not specify a valid \"DestTcpSvrAddrs\",it is virtually empty")
		flag.Usage()
		os.Exit(0)
	}
}

func main() {

	defer func() {
		//recover()
		fmt.Println("defer")
	}()

	panic("a")
}

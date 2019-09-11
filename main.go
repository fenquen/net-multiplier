package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"tcp-multiplier/config"
)

func main() {
	if "" == strings.Trim(config.DestTcpSvrAddrs, " ") {
		log.Println("you actually did not specify a valid \"DestTcpSvrAddrs\",it is virtually empty")
		flag.Usage()
		os.Exit(0)
	}

	
}


package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"tcp-multiplier/config"
	"tcp-multiplier/server"
)

func main() {
	defer func() {
		err := recover()
		if nil != err {
			log.Println(err)
		}
	}()

	//a:=make(chan int,1)
	//a<-0100
	//close(a)
	//fmt.Println(<-a)
	//fmt.Println(<-a)

	if "" == strings.Trim(config.DestTcpSvrAddrs, " ") {
		log.Println("you actually did not specify a valid \"DestTcpSvrAddrs\",it is virtually empty")
		flag.Usage()
		os.Exit(0)
	}

	server.ListenAndServe()
}

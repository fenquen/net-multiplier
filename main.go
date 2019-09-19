package main

import (
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

	if "" == strings.Trim(config.DestSvrAddrs, " ") {
		log.Println("you actually did not specify a valid \"DestSvrAddrs\",it is virtually empty")
		//flag.Usage()
		//os.Exit(0)
	}

	// verify mode
	switch config.Mode {
	case config.TCP_TYPE:
	case config.UDP_TYPE:
	default:
		log.Println("mode must be tcp or udp")
		os.Exit(0)
	}

	server.ListenAndServe()
}

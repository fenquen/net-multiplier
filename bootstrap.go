package main

import (
	"flag"
	"fmt"
	"net-multiplier/config"
	"net-multiplier/server"
	"net-multiplier/zaplog"
	"os"
	"strings"
)

func main() {
	defer func() {
		zaplog.LOGGER.Sync()
		err := recover()
		if nil != err {
			zaplog.LOGGER.Error(fmt.Sprint(err))
		}
	}()

	flag.Usage()

	if nil != flag.CommandLine.Lookup("h") {
		os.Exit(0)
	}

	if "" == strings.Trim(*config.DestSvrAddrs, " ") {
		zaplog.LOGGER.Info("you actually did not specify a valid \"DestSvrAddrs\",it is virtually empty")
		//flag.Usage()
		//os.Exit(0)
	}

	go server.ServeHttp()

	zaplog.LOGGER.Info("mode " + *config.Mode)

	// verify mode
	switch *config.Mode {
	case config.TCP_MODE:
		server.ListenAndServeTcp()
	case config.UDP_MODE:
		server.ServeUdp()
	default:
		zaplog.LOGGER.Info("mode must be tcp or udp")
		os.Exit(0)
	}

}

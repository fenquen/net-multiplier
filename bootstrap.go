package main

import (
	"fmt"
	"os"
	"strings"
	"tcp-multiplier/config"
	"tcp-multiplier/server"
	"tcp-multiplier/zaplog"
)

func main() {
	defer func() {
		zaplog.LOGGER.Sync()
		err := recover()
		if nil != err {
			zaplog.Error(fmt.Sprint(err))
		}
	}()

	if "" == strings.Trim(config.DestSvrAddrs, " ") {
		zaplog.Info("you actually did not specify a valid \"DestSvrAddrs\",it is virtually empty")
		//flag.Usage()
		//os.Exit(0)
	}

	zaplog.Info("mode " + config.Mode)

	// verify mode
	switch config.Mode {
	case config.TCP_MODE:
		server.ListenAndServeTcp()
	case config.UDP_MODE:
		server.ServeUdp()
	default:
		zaplog.Info("mode must be tcp or udp")
		os.Exit(0)
	}

}

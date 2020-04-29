package main

import (
	"flag"
	"fmt"
	"net-multiplier/config"
	"net-multiplier/server"
	"net-multiplier/zaplog"
	"os"
)

func main() {
	defer func() {
		_ = zaplog.LOGGER.Sync()
		err := recover()
		if nil != err {
			zaplog.LOGGER.Error(fmt.Sprint(err))
		}
	}()

	flag.Usage()

	if nil != flag.CommandLine.Lookup("h") {
		os.Exit(0)
	}

	/*if "" == strings.Trim(*config.DestSvrAddrs, " ") {
		zaplog.LOGGER.Info("you actually did not specify a valid \"DestSvrAddrs\" in config,it is virtually empty")
		//flag.Usage()
		//os.Exit(0)
	}*/
	zaplog.LOGGER.Info("defaultMode " + *config.DefaultMode)

	server.ServeHttp()

	// verify mode
	/*switch *config.DefaultMode {
	case config.TCP_MODE:
		server.listenAndServeTcp()
	case config.UDP_MODE:
		server.serveUdp()
	default:
		zaplog.LOGGER.Info("mode must be tcp or udp")
		os.Exit(0)
	}*/

}
